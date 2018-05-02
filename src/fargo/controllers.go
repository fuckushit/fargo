package fargo

import (
	"bdlib/logger"
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"bdlib/config"
	"bdlib/util"
	"fargo/context"
	"fargo/middleware"
)

// Controller 每一个路由的控制层对象, 将在应用的 controller 中被继承, 作为逻辑操作实体.
type Controller struct {
	// 上下文.
	Ctx *context.Context

	// config interface.
	Cfg config.Configer

	// controller 内容.
	Data map[interface{}]interface{}

	// controller 的名称.
	controllerName string

	// 每个 controller 对应的 action 名称.
	actionName string

	// 模板路径.
	TplNames string

	// 模板后缀之前的名称, 如 index.html 的 index.
	TplPrefix string

	// layout 路径.
	Layout string

	// 支持 layout 选项区域.
	LayoutSections map[string]string

	// 模板后缀, 如 .html, .tpl.
	TplExt string

	// xsrf 的 token 值
	_xsrfToken string

	// // 当前 session 操作对象, 通过这个对象可以进行 session Set、Get 等操作.
	// CruSession session.SessionStore

	// xsrf 的有效时间.
	XSRFExpire int64

	// app 的 controller.
	AppController interface{}

	// 是否开启自动渲染模板.
	EnableReander bool

	// 日志操作句柄.
	Loger *logger.Logger

	// 加入到accesslog中的日志
	Accesslog string
}

// ControllerInterface 控制层接口 每一个控制层结构体可以实现如下方法.
type ControllerInterface interface {
	Init(ct *context.Context, ControllerName, actionName string, app interface{})
	Prepare()
	Filter() bool
	Get()
	Post()
	Delete()
	Put()
	Head()
	Patch()
	Options()
	Finish()
	Render() error
	XsrfToken() string
	CheckXSRFCookie() bool
	accessLog(reqTime time.Duration, reqUnix int64)
}

// Init 控制层初始化.
// Parameters:
// - ctx:            上下文输入输出对象.
// - ControllerName: controller 的名称.
// - actionName:     controller 对应的 action 的名称.
// - app:            应用.
// Return:
//  - app:           fargo 对象.
func (c *Controller) Init(ctx *context.Context, ControllerName, actionName string, app interface{}) {
	c.Ctx = ctx
	c.Data = make(map[interface{}]interface{})
	c.controllerName = ControllerName
	c.actionName = actionName
	c.TplNames = ""
	c.TplPrefix = ""
	c.TplExt = "html"
	c.Layout = ""
	c.LayoutSections = make(map[string]string)
	c.XSRFExpire = XSRFExpire
	c.Data = ctx.Input.Data
	c.EnableReander = true
	c.AppController = app
	c.Loger = Log
	c.Cfg = gCfg
}

// Prepare 每一个请求到来之后处理到 Get、Post 等方法之前执行, 用于 ip 限制、黑白名单等限制工作.
func (c *Controller) Prepare() {

}

// Filter 过滤函数，根据返回true或者false决定接下来的操作
// 默认返回true
func (c *Controller) Filter() bool {
	return true
}

// Finish 每一个请求到来之后处理到 Get、Post 等方法之后执行, 用于 GC 等.
func (c *Controller) Finish() {

}

// accessLog 记录 access log.
func (c *Controller) accessLog(reqTime time.Duration, reqUnix int64) {
	// 客户端请求 ip.
	ip := c.Ctx.Input.IP()
	// http 请求状态.
	//HTTPStatus := c.Ctx.Output.Status
	// refer.
	refer := c.Ctx.Input.Refer()
	// host.
	host := c.Ctx.Input.Host()
	// system.
	//system := c.Ctx.Input.System()
	// browser.
	//browser := c.Ctx.Input.Browser()
	// user agent.
	userAgent := c.Ctx.Input.UserAgent()
	// language.
	//lang := c.Ctx.Input.Header("Accept-Language")
	// method.
	method := c.Ctx.Input.Method()
	// uRL.
	uRL := c.Ctx.Input.URL()

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "ACCESS_INFO: %s %s %s %s %s %s", ip, uRL, host, method, refer, userAgent)
	if c.Accesslog == "" {
		Log.PrintfN(2, "%s %s", buf.String(), reqTime.String())
	} else {
		Log.PrintfN(2, "%s %s %s", buf.String(), c.Accesslog, reqTime.String())
	}

	c.Data["fargo_req_time"] = reqTime.String()
	c.Data["fargo_req_now"] = reqUnix
}

// Get 将 GET 请求处理到对应的控制层上的 Get 方法上.
// 在应用重写之前, 默认是 405 页面.
func (c *Controller) Get() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Post 将 POST 请求处理到对应的控制层上的 Post 方法上.
// 在应用重写之前, 默认是 405 页面.
func (c *Controller) Post() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Delete 将 DELETE 请求处理到对应的控制层上的 Delete 方法上.
// 在应用重写之前, 默认是 405 页面.
func (c *Controller) Delete() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Put 将 PUT 请求处理到对应的控制层上的 Put 方法上.
// 在应用重写之前, 默认是 405 页面.
func (c *Controller) Put() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Head 将 HEAD 请求处理到对应的控制层上的 Head 方法上.
// 在应用重写之前, 默认是 405 页面.
func (c *Controller) Head() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Patch 将 PATCH 请求处理到对应的控制层上的 Patch 方法上.
// 在应用重写之前, 默认是 405 页面.
func (c *Controller) Patch() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Options 将 OPTIONS 请求处理到对应的控制层上的 Options 方法上.
// 在应用重写之前, 默认是 405 页面.
func (c *Controller) Options() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Render 将模板渲染到 response 中输出.
// Return:
// - err:
func (c *Controller) Render() (err error) {
	if !c.EnableReander {
		return
	}
	rb, err := c.RenderBytes()
	if err != nil {
		return
	}

	c.Ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
	c.Ctx.Output.Body(rb)

	return
}

// RenderString 将渲染的模板转换成字符串进行输出.
// Return:
// - tpl: 对应的 controller 渲染的模板转换成的字符串.
// - err:
func (c *Controller) RenderString() (tpl string, err error) {
	b, e := c.RenderBytes()
	return string(b), e
}

// RenderBytes 将渲染的模板转换成 byte 流进行输出.
// Return:
// - tpl: 对应的 controller 渲染的模板转换成的 byte 流.
// - err:
func (c *Controller) RenderBytes() (tpl []byte, err error) {
	// 如果 controller 设置了 layout, 则首先渲染出 layout.
	if c.Layout != "" {
		// 如果没有设置模板路径 则默认以 controller name 作为模板文件名.
		if c.TplNames == "" {
			c.TplNames = strings.ToLower(c.controllerName) + "/" + strings.ToLower(c.actionName) + "." + c.TplExt
		}
		// 如果设置了 debug 模式, 则每次请求都会重新编译模板文件,
		// 用于调试时更改模板文件后, 不需要重新编译运行程序.
		if gDebug {
			BuildTemplate(templateDirc)
		}

		newbytes := bytes.NewBufferString("")
		if _, ok := GetTemplates()[c.TplNames]; !ok {
			return []byte{}, fmt.Errorf("can't find templatefile in the path: %s", c.TplNames)
		}
		err := GetTemplates()[c.TplNames].ExecuteTemplate(newbytes, c.TplNames, c.Data)
		if err != nil {
			return nil, err
		}
		tplcontent, _ := ioutil.ReadAll(newbytes)
		c.Data["LayoutContent"] = template.HTML(string(tplcontent))

		// 支持 layout section.
		if c.LayoutSections != nil {
			for sectionName, sectionTpl := range c.LayoutSections {
				if sectionTpl == "" {
					c.Data[sectionName] = ""
					continue
				}
				sectionBytes := bytes.NewBufferString("")
				err = GetTemplates()[sectionTpl].ExecuteTemplate(sectionBytes, sectionTpl, c.Data)
				if err != nil {
					return nil, err
				}
				sectionContent, _ := ioutil.ReadAll(sectionBytes)
				c.Data[sectionName] = template.HTML(string(sectionContent))
			}
		}

		ibytes := bytes.NewBufferString("")
		err = GetTemplates()[c.Layout].ExecuteTemplate(ibytes, c.Layout, c.Data)
		if err != nil {
			return nil, err
		}
		icontent, _ := ioutil.ReadAll(ibytes)

		return icontent, nil
	}

	// 如果没有设置模板路径 则默认以 controller name 作为模板文件名.
	if c.TplNames == "" {
		c.TplNames = strings.ToLower(c.controllerName) + "/" + strings.ToLower(c.actionName) + "." + c.TplExt
	}
	// 如果设置了 debug 模式, 则每次请求都会重新编译模板文件,
	// 用于调试时更改模板文件后, 不需要重新编译运行程序.
	if gDebug {
		BuildTemplate(templateDirc)
	}

	ibytes := bytes.NewBufferString("")
	if _, ok := GetTemplates()[c.TplNames]; !ok {
		return []byte{}, fmt.Errorf("can't find templatefile in the path:%s", c.TplNames)
	}
	err = GetTemplates()[c.TplNames].ExecuteTemplate(ibytes, c.TplNames, c.Data)
	if err != nil {
		return nil, err
	}
	icontent, _ := ioutil.ReadAll(ibytes)

	return icontent, nil
}

// Redirect 通过给定的状态码(301, 302, etc.) 跳转到指定的 url.
// Parameters:
// - url:  跳转 url.
// - code: 跳转状态码, 301, 302 等.
func (c *Controller) Redirect(url string, code int) {
	c.Ctx.Redirect(code, url)
}

// Input 从 request 中获取输入的参数, 如表单数据, url 参数等.
// Return:
// - input: 输入的参数, 如表单数据, url 参数等.
func (c *Controller) Input() (input url.Values) {
	ct := c.Ctx.Request.Header.Get("Content-Type")
	if strings.Contains(ct, "multipart/form-data") {
		c.Ctx.Request.ParseMultipartForm(maxMemory) //64MB
	} else {
		c.Ctx.Request.ParseForm()
	}
	return c.Ctx.Request.Form
}

// GetString 从 request 中获取输入的参数, 如表单数据, url 参数等, key 值对应的值.
// Parameters:
// - key:   要获取的输入值的 key 值.
// Return:
// - value: 要获取的输入值的 key 值对应的值.
func (c *Controller) GetString(key string) (value string) {
	return c.Input().Get(key)
}

// GetStrings 从 request 中获取输入的参数数组,
// 应用于如 checkbox(input[type=chackbox]), 多选框等情况的表单提交.
// Parameters:
// - key:    要获取的输入值的 key 值.
// Return:
// - values: 要获取的输入值的 key 值对应的值数组.
func (c *Controller) GetStrings(key string) (values []string) {
	r := c.Ctx.Request
	if r.Form == nil {
		return
	}
	vs := r.Form[key]
	if len(vs) > 0 {
		return vs
	}

	return
}

// GetStringms 从 request 中获取输入的参数数组
// 应用于如 <input type="text" name="[1]name" value="吕超飞" class="x-input x-input-date"/>
//		   <input type="text" name="[3]name" value="张晔" class="x-input x-input-date"/>
//         <input type="text" name="[6]name" value="dogegg" class="x-input x-input-date"/>
//         <input type="text" name="[11]name" value="微微" class="x-input x-input-date"/>
func (c *Controller) GetStringms(key string) (values map[string][]string) {
	r := c.Ctx.Request
	if r.Form == nil {
		return
	}
	values = make(map[string][]string, 0)
	for k, v := range r.Form {
		ks := strings.Split(k, "]")
		if len(ks) == 2 && ks[1] == key {
			kss := strings.Split(ks[0], "[")
			if len(kss) == 2 {
				values[kss[1]] = v
			}
		}
	}
	return
}

// GetStringm _
func (c *Controller) GetStringm(key string) (values map[string]string) {
	r := c.Ctx.Request
	if r.Form == nil {
		return
	}
	values = make(map[string]string, 0)
	for k, v := range r.Form {
		ks := strings.Split(k, "]")
		if len(ks) == 2 && ks[1] == key {
			kss := strings.Split(ks[0], "[")
			if len(kss) == 2 {
				if len(v) > 0 {
					values[kss[1]] = v[0]
				}
			}
		}
	}
	return
}

// GetInt 从 url 参数中获得 int 变量值.
// Parameters:
// - key:   要获取的输入值的 key 值.
// Return:
// - value: 要获取的输入值的 key 值对应的值.
// - err:
func (c *Controller) GetInt(key string) (value int64, err error) {
	return strconv.ParseInt(c.Input().Get(key), 10, 64)
}

// GetBool 获得表单提交的 bool 类型变量值.
// Parameters:
// - key:   要获取的输入值的 key 值.
// Return:
// - value: 要获取的输入值的 key 值对应的值.
// - err:
func (c *Controller) GetBool(key string) (value bool, err error) {
	return strconv.ParseBool(c.Input().Get(key))
}

// GetFloat 获得 表单提交的 float64 类型变量值.
// Parameters:
// - key:   要获取的输入值的 key 值.
// Return:
// - value: 要获取的输入值的 key 值对应的值.
// - err:
func (c *Controller) GetFloat(key string) (value float64, err error) {
	return strconv.ParseFloat(c.Input().Get(key), 64)
}

// GetFile 获得上传的文件.
// Parameters:
// - key:    要获取的输入值的 key 值.
// Return:
// - file:   上传的文件.
// - header: 上传文件的头部信息.
// - err:
func (c *Controller) GetFile(key string) (file multipart.File, header *multipart.FileHeader, err error) {
	return c.Ctx.Request.FormFile(key)
}

// SaveToFile 将上传的文件转存成文件.
// Parameters:
// - fromfile: 上传的文件.
// - tofile:   转存的文件.
// Return:
// - err:
func (c *Controller) SaveToFile(fromfile, tofile string) (err error) {
	file, _, err := c.Ctx.Request.FormFile(fromfile)
	if err != nil {
		return
	}
	defer file.Close()
	f, err := os.OpenFile(tofile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return
	}
	defer f.Close()
	io.Copy(f, file)

	return
}

// TODO session

// IsAjax 判断这个请求是否为 ajax 请求.
func (c *Controller) IsAjax() (is bool) {
	return c.Ctx.Input.IsAjax()
}

// GetSecureCookie 获取解码之后的加密 cookie.
func (c *Controller) GetSecureCookie(secret, key string) (value string, ok bool) {
	val := c.Ctx.GetCookie(key)
	if val == "" {
		return
	}

	parts := strings.SplitN(val, "|", 3)
	if len(parts) != 3 {
		return
	}

	vs := parts[0]
	timestamp := parts[1]
	sig := parts[2]

	h := hmac.New(sha1.New, []byte(secret))
	fmt.Fprintf(h, "%s%s", vs, timestamp)

	if fmt.Sprintf("%02x", h.Sum(nil)) != sig {
		return
	}
	res, _ := base64.URLEncoding.DecodeString(vs)
	value = string(res)
	ok = true

	return
}

// SetSecureCookie 设置加密编码之后的 value 到 cookie.
func (c *Controller) SetSecureCookie(secret, name, value string, age int64) {
	vs := base64.URLEncoding.EncodeToString([]byte(value))
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	h := hmac.New(sha1.New, []byte(secret))
	fmt.Fprintf(h, "%s%s", vs, timestamp)
	sig := fmt.Sprintf("%02x", h.Sum(nil))

	cookie := strings.Join([]string{vs, timestamp, sig}, "|")
	c.Ctx.SetCookie(name, cookie, age, "/")

	return
}

// XsrfToken 生成 xsrf token.
func (c *Controller) XsrfToken() (token string) {
	if c._xsrfToken == "" {
		token, ok := c.GetSecureCookie(XSRFKEY, "_xsrf")
		if !ok {
			var expire int64
			if c.XSRFExpire > 0 {
				expire = int64(c.XSRFExpire)
			} else {
				expire = int64(XSRFExpire)
			}
			token = util.RandomString(15)
			c.SetSecureCookie(XSRFKEY, "_xsrf", token, expire)
		}
		c._xsrfToken = token
	}

	return c._xsrfToken
}

// CheckXSRFCookie 校验请求中的 xsrf token 是否合法,
// 可以在表单中的 "_xsrf" 和 header 中的 "X-Xsrftoken" 和 "X-CsrfToken" 获取.
func (c *Controller) CheckXSRFCookie() (ok bool) {
	token := c.GetString("_xsrf")
	if token == "" {
		token = c.Ctx.Request.Header.Get("X-Xsrftoken")
	}
	if token == "" {
		token = c.Ctx.Request.Header.Get("X-Csrftoken")
	}
	if token == "" {
		middleware.Exception("403", c.Ctx.ResponseWriter, c.Ctx.Request, "")
		return
	} else if c._xsrfToken != token {
		middleware.Exception("403", c.Ctx.ResponseWriter, c.Ctx.Request, "")
		return
	}

	return true
}

// XsrfFormHTML 生成 xsrf token input 的 html.
func (c *Controller) XsrfFormHTML() (xsrf string) {
	return fmt.Sprintf("<input type=\"hidden\" name=\"_xsrf\" value=\"%s\"/>", c._xsrfToken)
}

// GetConfiger 获取配置文件句柄.
func (c *Controller) GetConfiger() (cfg config.Configer) {
	return c.Cfg
}

// GetCfgSection 获取配置文件中的 section.
// Parameters:
// - sectionName: 要获取配置文件中的 section 名, 如配置文件中的 [database], 则此为 database.
// Return:
// - section:    配置文件中一个 section 的内容 map.
// - err:
func (c *Controller) GetCfgSection(sectionName string) (section config.Section, err error) {
	if c.Cfg == nil {
		return
	}

	return c.Cfg.GetSection(sectionName)
}

// GetCfgSetting 获取配置文件中某一个 setting 的值, 如 host=10.100.100.100, 则为 10.100.100.100.
// Parameters:
// - sectionName: 要获取配置文件中的 section 名, 如配置文件中的 [database], 则此为 database.
// - keyName:     要获得的设置的名称的名字, 如 host=10.100.100.100 中的 host.
// Return:
// - value:       某一个设置项的值.
// - err:
func (c *Controller) GetCfgSetting(sectionName, keyName string) (value string, err error) {
	if c.Cfg == nil {
		return
	}

	return gCfg.GetSetting(sectionName, keyName)
}

// GetCfgIntSetting 获取配置文件中一个整型的 setting 值.
// Parameters:
// - sectionName: 要获取配置文件中的 section 名, 如配置文件中的 [database], 则此为 database.
// - keyName:     要获得的设置的名称的名字, 如 host=10.100.100.100 中的 host.
// - dfault:      获取的整型变量不存在的时候的默认值.
// Return:
// - value:       某一个设置项的值.
// - err:
func (c *Controller) GetCfgIntSetting(sectionName, keyName string, dfault int64) (value int64, err error) {
	if c.Cfg == nil {
		return
	}

	return gCfg.GetIntSetting(sectionName, keyName, dfault)
}

// GetCfgBoolSetting 获取配置文件中一个 bool 的 setting 值.
// Parameters:
// - sectionName: 要获取配置文件中的 section 名, 如配置文件中的 [database], 则此为 database.
// - keyName:     要获得的设置的名称的名字, 如 host=10.100.100.100 中的 host.
// - dfault:      获取的 bool 变量不存在的时候的默认值.
// Return:
// - value:       某一个设置项的值.
// - err:
func (c *Controller) GetCfgBoolSetting(sectionName, keyName string, dfault bool) (value bool, err error) {
	if c.Cfg == nil {
		return
	}

	return gCfg.GetBoolSetting(sectionName, keyName, dfault)
}
