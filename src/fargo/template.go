package fargo

import (
	"bdlib/util"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	// fargoTplFuncMap 模板函数集合.
	fargoTplFuncMap = make(template.FuncMap)

	// FargoTemplates Fargo template 缓存.
	tmplMutex      sync.RWMutex
	FargoTemplates map[string]*template.Template

	// FargoTemplateExt 模板支持的后缀名 目前支持 tpl, html 两种.
	FargoTemplateExt []string
)

func GetTemplates() map[string]*template.Template {
	if !gDebug {
		return FargoTemplates
	}

	tmplMutex.RLock()
	v := FargoTemplates
	tmplMutex.RUnlock()

	return v
}

// init 初始化 Fargo 模板对象.
func init() {
	FargoTemplates = make(map[string]*template.Template)
	FargoTemplateExt = make([]string, 0)
	FargoTemplateExt = append(FargoTemplateExt, "tpl", "html")
}

// AddFuncMap 用户可以添加模板函数.
func AddFuncMap(key string, funcname interface{}) (err error) {
	fargoTplFuncMap[key] = funcname

	return
}

// templatefile 模板文件对象.
type templatefile struct {
	// 根目录.
	root string

	// 目录下的模板文件.
	files map[string][]string
}

// visit 一层一层目录的去查找模板文件.
// Parameters:
// - paths: 路径.
// - f:     文件句柄.
// - err:   错误.
// Return:
func (t *templatefile) visit(paths string, f os.FileInfo) (err error) {
	if f == nil {
		return nil
	}
	// 文件是文件夹或者软链接.
	if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
		return nil
	}
	// 后缀名不支持.
	if !HasTemplateExt(paths) {
		return nil
	}

	// 将路径 "\" 替换成 "/", 用于 windows 下面的路径替换.
	replace := strings.NewReplacer("\\", "/")
	a := []byte(paths)
	a = a[len([]byte(t.root)):]
	file := strings.TrimLeft(replace.Replace(string(a)), "/")
	subdir := filepath.Dir(file)
	if _, ok := t.files[subdir]; ok {
		t.files[subdir] = append(t.files[subdir], file)
	} else {
		m := make([]string, 1)
		m[0] = file
		t.files[subdir] = m
	}

	return nil
}

// HasTemplateExt paths 下是否包含 Fargo 支持的 模板后缀名.
// Parameters:
// - paths: http 请求的 URI, 如 /css, /js 等.
// Return:
// - has:   是否包含.
func HasTemplateExt(paths string) (has bool) {
	for _, v := range FargoTemplateExt {
		if strings.HasSuffix(paths, fmt.Sprintf(".%s", v)) {
			return true
		}
	}

	return
}

// AddTemplateExt 添加支持的新的模板扩展, Fargo 框架模板默认支持的模板文件后缀为 html 和 tpl 两种.
// Parameters:
// - ext: 要添加支持的模板后缀.
func AddTemplateExt(ext string) {
	for _, v := range FargoTemplateExt {
		if v == ext {
			return
		}
	}
	FargoTemplateExt = append(FargoTemplateExt, ext)
}

// BuildTemplate 渲染文件夹下面的所有模板文件, Fargo 框架采用预编译模板文件的模式,
// 在应用运行的时候就会一次性编译所有的模板文件到模板缓存中.
// Parameters:
// - dir: 模板文件目录.
// Return:
// - err:
func BuildTemplate(dir string) (err error) {

	if _, err = os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("dir open err")
	}
	// 初始化文件模板.
	tf := &templatefile{
		root:  dir,
		files: make(map[string][]string),
	}
	// 从根目录遍历模板文件.
	err = filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return tf.visit(path, f)
	})
	if err != nil {
		return
	}

	tmpl := make(map[string]*template.Template)

	for _, v := range tf.files {
		for _, file := range v {
			t, err := getTemplate(tf.root, file, v...)
			if err != nil {
				fmt.Println(err)
				continue
			} else {
				tmpl[file] = t
			}
		}
	}

	tmplMutex.Lock()
	FargoTemplates = tmpl
	tmplMutex.Unlock()

	return
}

// getTplDeep 获取模板文件目录深度.
// Parameters:
// - root: 模板根目录.
// - file: 要获取的模板文件名称.
// - t:    编译的模板文件对象.
// Return；
// 编译的模板文件.
// 查找匹配的所有结果集.
// 错误.
func getTplDeep(root, file, parent string, t *template.Template) (*template.Template, [][]string, error) {
	var fileabspath string
	if filepath.HasPrefix(file, "../") {
		fileabspath = filepath.Join(root, filepath.Dir(parent), file)
	} else {
		fileabspath = filepath.Join(root, file)
	}
	if e := util.FileExists(fileabspath); !e {
		return nil, [][]string{}, fmt.Errorf("can't find template file %s", file)
	}
	data, err := ioutil.ReadFile(fileabspath)
	if err != nil {
		return nil, [][]string{}, err
	}
	t, err = t.New(file).Parse(string(data))
	if err != nil {
		return nil, [][]string{}, err
	}
	reg := regexp.MustCompile(gTemplateLeft + "[ ]*template[ ]+\"([^\"]+)\"")
	allsub := reg.FindAllStringSubmatch(string(data), -1)
	for _, m := range allsub {
		if len(m) == 2 {
			tlook := t.Lookup(m[1])
			if tlook != nil {
				continue
			}
			if !HasTemplateExt(m[1]) {
				continue
			}
			t, _, err = getTplDeep(root, m[1], file, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}

	return t, allsub, nil
}

// getTemplate 获取模板文件.
func getTemplate(root, file string, others ...string) (t *template.Template, err error) {
	t = template.New(file).Delims(gTemplateLeft, gTemplateRight).Funcs(fargoTplFuncMap)
	var submods [][]string
	t, submods, err = getTplDeep(root, file, "", t)
	if err != nil {
		return
	}
	t, err = _getTemplate(t, root, submods, others...)
	if err != nil {
		return
	}

	return
}

// _getTemplate 私有的获取文件.
func _getTemplate(t0 *template.Template, root string, submods [][]string, others ...string) (t *template.Template, err error) {
	t = t0
	for _, m := range submods {
		if len(m) == 2 {
			templ := t.Lookup(m[1])
			if templ != nil {
				continue
			}
			// 首先检测文件名
			for _, otherfile := range others {
				if otherfile == m[1] {
					var submods1 [][]string
					t, submods1, err = getTplDeep(root, otherfile, "", t)
					if err != nil {
						continue
					} else if submods1 != nil && len(submods1) > 0 {
						t, err = _getTemplate(t, root, submods1, others...)
					}
					break
				}
			}
			// 检测定义
			for _, otherfile := range others {
				fileabspath := filepath.Join(root, otherfile)
				data, err := ioutil.ReadFile(fileabspath)
				if err != nil {
					continue
				}
				reg := regexp.MustCompile(gTemplateLeft + "[ ]*define[ ]+\"([^\"]+)\"")
				allsub := reg.FindAllStringSubmatch(string(data), -1)
				for _, sub := range allsub {
					if len(sub) == 2 && sub[1] == m[1] {
						var submods1 [][]string
						t, submods1, err = getTplDeep(root, otherfile, "", t)
						if err != nil {
							continue
						} else if submods1 != nil && len(submods1) > 0 {
							t, err = _getTemplate(t, root, submods1, others...)
						}
						break
					}
				}
			}
		}
	}

	return
}
