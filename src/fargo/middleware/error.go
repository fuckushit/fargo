package middleware

import (
	"bdlib/comm"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

// some global vars.
var (
	AppName string
	VERSION string
)

var tpl = `
<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    <title>fargo application error</title>
    <style>
        html, body, body * {padding: 0; margin: 0;}
        #header {background:#ffd; border-bottom:solid 2px #A31515; padding: 20px 10px;}
        #header h2{ }
        #footer {border-top:solid 1px #aaa; padding: 5px 10px; font-size: 12px; color:green;}
        #content {padding: 5px;}
        #content .stack b{ font-size: 13px; color: red;}
        #content .stack pre{padding-left: 10px;}
        table {}
        td.t {text-align: right; padding-right: 5px; color: #888;}
    </style>
    <script type="text/javascript">
    </script>
</head>
<body>
    <div id="header">
        <h2>{{.AppError}}</h2>
    </div>
    <div id="content">
        <table>
            <tr>
                <td class="t">Request Method: </td><td>{{.RequestMethod}}</td>
            </tr>
            <tr>
                <td class="t">Request URL: </td><td>{{.RequestURL}}</td>
            </tr>
            <tr>
                <td class="t">RemoteAddr: </td><td>{{.RemoteAddr }}</td>
            </tr>
        </table>
        <div class="stack">
            <b>Stack</b>
            <pre>{{.Stack}}</pre>
        </div>
    </div>
    <div id="footer">
        <p>fargo {{ .fargoVersion }} (fargo framework)</p>
        <p>golang version: {{.GoVersion}}</p>
    </div>
</body>
</html>
`

// ShowErr 渲染默认的应用错误页面
func ShowErr(err interface{}, rw http.ResponseWriter, r *http.Request, stack string) {
	t, _ := template.New("fargoerrortemp").Parse(tpl)
	data := make(map[string]string)
	data["AppError"] = AppName + ":" + fmt.Sprint(err)
	data["RequestMethod"] = r.Method
	data["RequestURL"] = r.RequestURI
	data["RemoteAddr"] = r.RemoteAddr
	data["Stack"] = stack
	data["fargoVersion"] = VERSION
	data["GoVersion"] = runtime.Version()
	rw.WriteHeader(500)
	t.Execute(rw, data)
}

// 错误页面模板
var errtpl = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="">
    <meta name="author" content="">
    <title>{{.Title}}</title>
    <style type="text/css">
    body { font-family: "Helvetica Neue", Helvetica, Arial, sans-serif; font-size: 14px; line-height: 1.42857143; color: #333; background-color: #fff; } 
    h1, h2, h3 {font-family: inherit; font-weight: 500; line-height: 1.1; color: inherit; font-family: 'Open Sans', sans-serif; font-weight: 300; margin-top: 20px; margin-bottom: 10px; }
    h1 {font-size: 36px; }
    h2 {font-size: 30px; }
    h3 {font-size: 24px; }
    .text-center {text-align: center; }
    a {color: #23c0a2; text-decoration: none; outline: 0 none; }
    a:focus, a:hover, a:active {outline: 0 none; text-decoration: none; color: #0fac8e; }
    #cl-wrapper {display: table; width: 100%; position: absolute; height: 100%; }
    .page-error {margin-top: 80px; margin-bottom: 40px; }
    .page-error .number {color: #FFF; font-size: 150px; font-family: Arial; text-shadow: 1px 1px 5px rgba(0, 0, 0, 0.6); }
    .page-error .description {color: #FFF; font-size: 40px; text-shadow: 1px 1px 5px rgba(0, 0, 0, 0.6); }
    .page-error h3 {color: #FFF; text-shadow: 1px 1px 5px rgba(0, 0, 0, 0.6); }
    .error-container .copy, .error-container .copy a {color: #C9D4F6; text-shadow: 1px 1px 0 rgba(0, 0, 0, 0.3); }
    body.texture {background: #23262b; } </style>
</head>
<body class="texture">
    <div id="cl-wrapper" class="error-container">
        <div class="page-error">
            <h1 class="number text-center">{{.Title}}</h1>
            <h2 class="description text-center">{{.Content}}</h2>
            <h3 class="text-center">Would you like to go <a href="/">home</a>?</h3>
        </div>
        <div class="text-center copy">&copy; 2015 <a href="https://www.baidu.com">www.baidu.com</a></div>
    </div>
</body>
</html>
`

// handleFunc 集合.
var ErrorMaps = make(map[string]http.HandlerFunc)

// NotFound 404 页面 - 文件未找到
func NotFound(rw http.ResponseWriter, r *http.Request) {
	t, _ := template.New("fargoerrortemp").Parse(errtpl)
	data := make(map[string]interface{})
	data["Title"] = "404"
	data["Content"] = template.HTML("Sorry, but this page doesn't exists!")
	data["fargoVersion"] = VERSION
	t.Execute(rw, data)
}

// Unauthorized 401 页面 - 未经授权
func Unauthorized(rw http.ResponseWriter, r *http.Request) {
	t, _ := template.New("fargoerrortemp").Parse(errtpl)
	data := make(map[string]interface{})
	data["Title"] = "401"
	data["Content"] = template.HTML("Sorry, but this page does Unauthorized!")
	data["fargoVersion"] = VERSION
	t.Execute(rw, data)
}

// Forbidden 403 页面 - 禁止访问
func Forbidden(rw http.ResponseWriter, r *http.Request) {
	t, _ := template.New("fargoerrortemp").Parse(errtpl)
	data := make(map[string]interface{})
	data["Title"] = "403"
	data["Content"] = template.HTML("Sorry, but this page does Forbidden!")
	data["fargoVersion"] = VERSION
	t.Execute(rw, data)
}

// ServiceUnavailable 503 页面 - 服务不可用
func ServiceUnavailable(rw http.ResponseWriter, r *http.Request) {
	t, _ := template.New("fargoerrortemp").Parse(errtpl)
	data := make(map[string]interface{})
	data["Title"] = "503"
	data["Content"] = template.HTML("Sorry, Service Unavailable!")
	data["fargoVersion"] = VERSION
	t.Execute(rw, data)
}

// InternalServerError 500 页面 - 服务器内部错误
func InternalServerError(rw http.ResponseWriter, r *http.Request) {
	t, _ := template.New("fargoerrortemp").Parse(errtpl)
	data := make(map[string]interface{})
	data["Title"] = "500"
	data["Content"] = template.HTML("Sorry, Internal Server Error!")
	data["fargoVersion"] = VERSION
	t.Execute(rw, data)
}

// SimpleServerError golang 原生错误提醒的 500 页面
func SimpleServerError(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// ErrorHandler 添加 错误的 handler
func ErrorHandler(err string, h http.HandlerFunc) {
	ErrorMaps[err] = h
}

// RegisterErrorHandler 注册默认的错误 handlers, 404, 401, 403, 500, 503
func RegisterErrorHandler(tplFile string, errTplFile string) {

	if tplFile != "" {
		f, err := os.Open(tplFile)
		if err != nil {
			fmt.Println(comm.WrapError(err))
			os.Exit(1)
		}
		defer f.Close()
		tplBytes, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println(comm.WrapError(err))
			os.Exit(1)
		}
		tpl = string(tplBytes)
	}
	if errTplFile != "" {
		f, err := os.Open(errTplFile)
		if err != nil {
			fmt.Println(comm.WrapError(err))
			os.Exit(1)
		}
		defer f.Close()
		tplBytes, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println(comm.WrapError(err))
			os.Exit(1)
		}
		errtpl = string(tplBytes)
	}

	// 404
	if _, ok := ErrorMaps["404"]; !ok {
		ErrorMaps["404"] = NotFound
	}
	// 401
	if _, ok := ErrorMaps["401"]; !ok {
		ErrorMaps["401"] = Unauthorized
	}
	// 403
	if _, ok := ErrorMaps["403"]; !ok {
		ErrorMaps["403"] = Forbidden
	}
	// 500
	if _, ok := ErrorMaps["500"]; !ok {
		ErrorMaps["500"] = InternalServerError
	}
	// 503
	if _, ok := ErrorMaps["503"]; !ok {
		ErrorMaps["503"] = ServiceUnavailable
	}
}

// Exception 将 err 作为 简洁提示消息
// 当 err 为空时, 展示 500 的错误作为默认
func Exception(errcode string, w http.ResponseWriter, r *http.Request, msg string) {
	if h, ok := ErrorMaps[errcode]; ok {
		isint, err := strconv.Atoi(errcode)
		if err != nil {
			isint = 500
		}
		w.WriteHeader(isint)
		h(w, r)
		return
	}

	isint, err := strconv.Atoi(errcode)
	if err != nil {
		isint = 500
	}
	if 400 == isint {
		msg = "404 page not found"
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(isint)
	fmt.Fprintln(w, msg)

	return
}
