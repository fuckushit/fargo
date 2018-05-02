package context

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pquerna/ffjson/ffjson"
)

// FargoOutput 输出封装.
type FargoOutput struct {
	Context    *Context
	Status     int
	EnableGzip bool
}

// NewOutput 新建一个 fargo 输出对象.
// Return:
//  - m: fargo 输出对象.
func NewOutput() (m *FargoOutput) {
	return new(FargoOutput)
}

// Header 设置输出的 header 信息, 例如 Header("Server", "fargo").
// Parameters:
// - key: header key 值, 如 server.
// - val: 值.
func (m *FargoOutput) Header(key, val string) {
	m.Context.ResponseWriter.Header().Set(key, val)
}

// Body 设置输出的内容信息, 例如 Body([]byte("baidu")).
// Return:
//  - content: 输出的内容信息.
func (m *FargoOutput) Body(content []byte) {
	outputWriter := m.Context.ResponseWriter.(io.Writer)
	if m.EnableGzip && m.Context.Input.Header("Accept-Encoding") != "" {
		splitted := strings.SplitN(m.Context.Input.Header("Accept-Encoding"), ",", -1)
		encodings := make([]string, len(splitted))

		for i, val := range splitted {
			encodings[i] = strings.TrimSpace(val)
		}
		for _, val := range encodings {
			if val == "gzip" {
				m.Header("Content-Encoding", "gzip")
				outputWriter, _ = gzip.NewWriterLevel(m.Context.ResponseWriter, gzip.BestSpeed)

				break
			} else if val == "deflate" {
				m.Header("Content-Encoding", "deflate")
				outputWriter, _ = flate.NewWriter(m.Context.ResponseWriter, flate.BestSpeed)

				break
			}
		}
	} else {
		m.Header("Content-Length", strconv.Itoa(len(content)))
	}
	outputWriter.Write(content)
	switch outputWriter.(type) {
	case *gzip.Writer:
		outputWriter.(*gzip.Writer).Close()
	case *flate.Writer:
		outputWriter.(*flate.Writer).Close()
	}
}

// Cookie 设置输出的 cookie 信息, 例如 Cookie("sessionID","fargoSessionID").
// Parameters:
// - name:   要设置的 cookie 的 key.
// - value:  要设置的 cookie 的 值.
// - others: 其他 cookie 的选项, 如 path、HttpOnly 等.
func (m *FargoOutput) Cookie(name string, value string, others ...interface{}) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s=%s", sanitizeName(name), sanitizeValue(value))
	if len(others) > 0 {
		switch others[0].(type) {
		case int:
			if others[0].(int) > 0 {
				fmt.Fprintf(&b, "; Max-Age=%d", others[0].(int))
			} else if others[0].(int) < 0 {
				fmt.Fprintf(&b, "; Max-Age=0")
			}
		case int64:
			if others[0].(int64) > 0 {
				fmt.Fprintf(&b, "; Max-Age=%d", others[0].(int64))
			} else if others[0].(int64) < 0 {
				fmt.Fprintf(&b, "; Max-Age=0")
			}
		case int32:
			if others[0].(int32) > 0 {
				fmt.Fprintf(&b, "; Max-Age=%d", others[0].(int32))
			} else if others[0].(int32) < 0 {
				fmt.Fprintf(&b, "; Max-Age=0")
			}
		}
	}
	if len(others) > 1 {
		fmt.Fprintf(&b, "; Path=%s", sanitizeValue(others[1].(string)))
	}
	if len(others) > 2 {
		fmt.Fprintf(&b, "; Domain=%s", sanitizeValue(others[2].(string)))
	}
	if len(others) > 3 {
		fmt.Fprintf(&b, "; Secure")
	}
	if len(others) > 4 {
		fmt.Fprintf(&b, "; HttpOnly")
	}
	m.Context.ResponseWriter.Header().Add("Set-Cookie", b.String())
}

var cookieNameSanitizer = strings.NewReplacer("\n", "-", "\r", "-")

// sanitizeName ...
func sanitizeName(n string) (sanitizer string) {
	return cookieNameSanitizer.Replace(n)
}

var cookieValueSanitizer = strings.NewReplacer("\n", " ", "\r", " ", ";", " ")

// sanitizeName ...
func sanitizeValue(v string) (sanitizer string) {
	return cookieValueSanitizer.Replace(v)
}

// JSON 把 Data 格式化为 Json, 然后调用 Body 输出数据.
// Parameters:
// - data:       要输出的数据.
// - hasIndent:  marshal 的 时候是否需要 Indet.
// - coding:     是否需要进行转码, 如 \\00 格式转换成字符串.
func (m *FargoOutput) JSON(data interface{}, hasIndent bool, coding bool) (err error) {
	m.Header("Content-Type", "application/json; charset=utf-8")
	var content []byte
	if hasIndent {
		content, err = json.MarshalIndent(data, "", "  ")
	} else {
		content, err = ffjson.Marshal(data)
	}
	if err != nil {
		http.Error(m.Context.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	if coding {
		content = []byte(stringsToJSON(string(content)))
	}
	m.Body(content)

	return
}

// Jsonp 把 Data 格式化为 Jsonp, 然后调用 Body 输出数据.
// Parameters:
// - data:       要输出的数据.
// - hasIndent:  marshal 的 时候是否需要 Indet.
func (m *FargoOutput) Jsonp(data interface{}, hasIndent bool) (err error) {
	m.Header("Content-Type", "application/javascript;charset=UTF-8")
	var content []byte
	if hasIndent {
		content, err = json.MarshalIndent(data, "", "  ")
	} else {
		content, err = ffjson.Marshal(data)
	}
	if err != nil {
		http.Error(m.Context.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	callback := m.Context.Input.Query("callback")
	if callback == "" {
		return errors.New(`"callback" parameter required`)
	}
	callbackContent := bytes.NewBufferString(" " + template.JSEscapeString(callback))
	callbackContent.WriteString("(")
	callbackContent.Write(content)
	callbackContent.WriteString(");\r\n")
	m.Body(callbackContent.Bytes())

	return
}

// XML 把 Data 格式化为 XML, 然后调用 Body 输出数据.
// Parameters:
// - data:       要输出的数据.
// - hasIndent:  marshal 的 时候是否需要 Indet.
func (m *FargoOutput) XML(data interface{}, hasIndent bool) (err error) {
	m.Header("Content-Type", "application/xml;charset=UTF-8")
	var content []byte
	if hasIndent {
		content, err = xml.MarshalIndent(data, "", "  ")
	} else {
		content, err = xml.Marshal(data)
	}
	if err != nil {
		http.Error(m.Context.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	m.Body(content)

	return
}

// Download 把 file 路径传递进来, 然后输出文件给用户.
// Parameters:
// - file:  文件路径.
func (m *FargoOutput) Download(file string) {
	m.Header("Content-Description", "File Transfer")
	m.Header("Content-Type", "application/octet-stream")
	m.Header("Content-Disposition", "attachment; filename="+filepath.Base(file))
	m.Header("Content-Transfer-Encoding", "binary")
	m.Header("Expires", "0")
	m.Header("Cache-Control", "must-revalidate")
	m.Header("Pragma", "public")
	http.ServeFile(m.Context.ResponseWriter, m.Context.Request, file)
}

// ContentType 设置输出的 ContentType.
// Parameters:
// - ext:  输出的类型.
func (m *FargoOutput) ContentType(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	ctype := mime.TypeByExtension(ext)
	if ctype != "" {
		m.Header("Content-Type", ctype)
	}
}

// SetStatus 设置输出的状态码, 如 404.
// Parameters:
// - status:  状态码.
func (m *FargoOutput) SetStatus(status int) {
	m.Context.ResponseWriter.WriteHeader(status)
	m.Status = status
}

// IsCachable 根据 status 判断，是否为缓存类的状态, 200, 300, 304 则为可缓存状态.
// Parameters:
// - status:  状态码.
// Return:
//  - is:     是否可以缓存.
func (m *FargoOutput) IsCachable(status int) (is bool) {
	return m.Status >= 200 && m.Status < 300 || m.Status == 304
}

// IsEmpty 根据 status 判断，是否为空的状态, 201, 204, 304 则为内容为空的状态.
// Parameters:
// - status:  状态码.
// Return:
//  - is:     是否为空.
func (m *FargoOutput) IsEmpty(status int) (is bool) {
	return m.Status == 201 || m.Status == 204 || m.Status == 304
}

// IsOk 根据 status 判断，是否为 200 的状态.
// Parameters:
// - status:  状态码.
// Return:
//  - is:     是否为200.
func (m *FargoOutput) IsOk(status int) (is bool) {
	return m.Status == 200
}

// IsSuccessful 根据 status 判断，是否为正常的状态, 为 2xx 状态.
// Parameters:
// - status:  状态码.
// Return:
//  - is:     是否正常.
func (m *FargoOutput) IsSuccessful(status int) (is bool) {
	return m.Status >= 200 && m.Status < 300
}

// IsRedirect 根据 status 判断，是否为跳转的状态, 为 301, 302, 303 或者 307.
// Parameters:
// - status:  状态码.
// Return:
//  - is:     是否跳转.
func (m *FargoOutput) IsRedirect(status int) (is bool) {
	return m.Status == 301 || m.Status == 302 || m.Status == 303 || m.Status == 307
}

// IsForbidden 根据 status 判断，是否为403.
// Parameters:
// - status:  状态码.
// Return:
//  - is:     是否403.
func (m *FargoOutput) IsForbidden(status int) (is bool) {
	return m.Status == 403
}

// IsNotFound 根据 status 判断，是否为404.
// Parameters:
// - status:  状态码.
// Return:
//  - is:     是否404.
func (m *FargoOutput) IsNotFound(status int) (is bool) {
	return m.Status == 404
}

// IsClientError 根据 status 判断，客户端是否出现错误, 如 4xx, 5xx.
// Parameters:
// - status:  状态码.
// Return:
//  - is:     是否客户端是否出现错误.
func (m *FargoOutput) IsClientError(status int) (is bool) {
	return m.Status >= 400 && m.Status < 600
}

// stringsToJson 将 json string 类型转换成 json 类型, 即将含有 /x33 的转换成 33 类型的.
func stringsToJSON(str string) string {
	rs := []rune(str)
	jsons := ""
	for _, r := range rs {
		rint := int(r)
		if rint < 128 {
			jsons += string(r)
		} else {
			jsons += "\\u" + strconv.FormatInt(int64(rint), 16) // json
		}
	}
	return jsons
}

// Session 设置在服务器端保存的值, 例如 Session("username","baidu"), 这样用户就可以在下次使用的时候读取.
// Parameters:
// - name:  要获取的 session 的 key.
// Return:
// - value: 要获取的 session 值.
func (m *FargoOutput) Session(name interface{}, value interface{}) {
	m.Context.Input.CruSession.Set(name, value)
}
