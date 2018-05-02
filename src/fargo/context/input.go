package context

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"fargo/session"
)

// FargoInput http 输入封装, request header, data, cookie, body 等操作,
// 同时含有 session.
type FargoInput struct {
	// session 句柄
	CruSession session.SessionStore

	// 参数
	Params map[string]string

	// 在控制层中调用的时候存储的数据
	Data map[interface{}]interface{}

	// http.Request
	Request *http.Request

	// request body
	RequestBody []byte
}

// NewInput 新建 Fargo context 输入操作对象.
// Return:
//  - m: Fargo context 输入操作对象.
func NewInput(r *http.Request) (m *FargoInput) {
	return &FargoInput{
		Params:  make(map[string]string),
		Data:    make(map[interface{}]interface{}),
		Request: r,
	}
}

// Protocol 获得使用的 HTTP 协议, 如 HTTP/1.1.
// Return:
//  - protocol: 协议.
func (m *FargoInput) Protocol() (protocol string) {
	return m.Request.Proto
}

// Uri 获得 http 请求的 URI, 如 /index, /hello 等.
// Return:
//  - uri: URI.
func (m *FargoInput) Uri() (uri string) {
	return m.Request.RequestURI
}

// URL 返回 http 请求的完整 url, 如 https://www.baidu.com.
// Return:
//  - url: URL.
func (m *FargoInput) URL() (url string) {
	return m.Request.URL.String()
}

// Site 请求的站点地址, scheme+doamin 的组合, 例如 https://www.baidu.com
// Return:
//  - site: 请求的站点 url.
func (m *FargoInput) Site() (site string) {
	return m.Scheme() + "://" + m.Domain()
}

// Scheme 请求的 http 传输协议, 例如 “http” 或者 “https”.
// Return:
//  - scheme: http 传输协议.
func (m *FargoInput) Scheme() (scheme string) {
	if m.Request.URL.Scheme != "" {
		return m.Request.URL.Scheme
	} else if m.Request.TLS == nil {
		return "http"
	} else {
		return "https"
	}
}

// Domain 请求的域名, 例如 www.baidu.com.
// Return:
//  - domain: 请求的域名.
func (m *FargoInput) Domain() (domain string) {
	return m.Host()
}

// Host 请求的域名, 和 domain 一样.
// Return:
//  - host: 请求的域名.
func (m *FargoInput) Host() (host string) {
	if m.Request.Host != "" {
		hostParts := strings.Split(m.Request.Host, ":")
		if len(hostParts) > 0 {
			return hostParts[0]
		}
		return m.Request.Host
	}

	return "localhost"
}

// Method 请求的方法, 标准的 HTTP 请求方法, 例如 GET、POST 等.
// Return:
//  - method: 请求的方法.
func (m *FargoInput) Method() (method string) {
	return m.Request.Method
}

// Is 判断是否是某一个方法, 例如 Is("GET") 返回 true.
// Return:
//  - is: 是否是某一个方法.
func (m *FargoInput) Is(method string) (is bool) {
	return m.Method() == method
}

// IsAjax 判断是否是 AJAX 请求, 如果是返回 true, 不是返回 false.
// Return:
//  - is: 是否是 AJAX 请求.
func (m *FargoInput) IsAjax() (is bool) {
	return m.Header("X-Requested-With") == "XMLHttpRequest"
}

// IsSecure 判断当前请求是否 HTTPS 请求, 是返回 true, 否返回 false.
// Return:
//  - is: 是否 HTTPS 请求.
func (m *FargoInput) IsSecure() (is bool) {
	return m.Scheme() == "https"
}

// IsWebsocket 判断当前请求是否 Websocket 请求, 如果是返回 true, 否返回 false.
// Return:
//  - is: 是否 Websocket 请求.
func (m *FargoInput) IsWebsocket() (is bool) {
	return m.Header("Upgrade") == "websocket"
}

// IsUpload 判断当前请求是否有文件上传, 有返回 true, 否返回 false
// Return:
//  - is: 是否有文件上传.
func (m *FargoInput) IsUpload() (is bool) {
	return m.Request.MultipartForm != nil
}

// IP 返回请求用户的 IP, 如果用户通过代理, 一层一层剥离获取真实的 IP.
// Return:
//  - ip: ip 地址, string 类型.
func (m *FargoInput) IP() (ip string) {
	ips := m.Proxy()
	if len(ips) > 0 && ips[0] != "" {
		return ips[0]
	}
	ipS := strings.Split(m.Request.RemoteAddr, ":")
	if len(ipS) > 0 {
		if ipS[0] != "[" {
			return ipS[0]
		}
	}

	return "127.0.0.1"
}

// Proxy 返回用户代理请求的所有 IP.
// Return:
//  - proxy: 所有代理集合.
func (m *FargoInput) Proxy() (proxy []string) {
	if ips := m.Header("X-Forwarded-For"); ips != "" {
		return strings.Split(ips, ",")
	}

	return []string{}
}

// Browser 客户端浏览器的类型.
// Return:
//  - browser: 浏览器.
func (m *FargoInput) Browser() (browser string) {
	userAgent := m.UserAgent()
	if userAgent == "" {
		return "机器人"
	} else if strings.Index(userAgent, "MSIE 9.0") != -1 {
		return "Internet Explorer 9.0"
	} else if strings.Index(userAgent, "MSIE 8.0") != -1 {
		return "Internet Explorer 8.0"
	} else if strings.Index(userAgent, "MSIE 7.0") != -1 {
		return "Internet Explorer 7.0"
	} else if strings.Index(userAgent, "MSIE 6.0") != -1 {
		return "Internet Explorer 6.0"
	} else if strings.Index(userAgent, "Firefox") != -1 {
		return "Firefox"
	} else if strings.Index(userAgent, "Chrome") != -1 {
		return "Chrome"
	} else if strings.Index(userAgent, "Safari") != -1 {
		return "Safari"
	} else if strings.Index(userAgent, "Opera") != -1 {
		return "Opera"
	} else if strings.Index(userAgent, "360SE") != -1 {
		return "360SE"
	} else {
		return "unknown"
	}
}

// System 客户端操作系统类型.
// Return:
//  - system: 操作系统.
func (m *FargoInput) System() (system string) {
	userAgent := m.UserAgent()
	if userAgent == "" {
		return "未知操作系统"
	} else if strings.Index(userAgent, "NT 6.1") != -1 {
		return "Windows 7"
	} else if strings.Index(userAgent, "NT 6.0") != -1 {
		return "Windows Vista"
	} else if strings.Index(userAgent, "NT 5.1") != -1 {
		return "Windows XP"
	} else if strings.Index(userAgent, "NT 5.2") != -1 {
		return "Windows Server 2003"
	} else if strings.Index(userAgent, "NT 5") != -1 {
		return "Windows 2000"
	} else if strings.Index(userAgent, "NT 4.9") != -1 {
		return "Windows ME"
	} else if strings.Index(userAgent, "NT 4") != -1 {
		return "Windows NT 4.0"
	} else if strings.Index(userAgent, "98") != -1 {
		return "Windows 98"
	} else if strings.Index(userAgent, "95") != -1 {
		return "Windows 95"
	} else if strings.Index(userAgent, "Mac") != -1 {
		return "Mac"
	} else if strings.Index(userAgent, "Linux") != -1 {
		return "Linux"
	} else if strings.Index(userAgent, "Unix") != -1 {
		return "Unix"
	} else if strings.Index(userAgent, "FreeBSD") != -1 {
		return "FreeBSD"
	} else if strings.Index(userAgent, "SunOS") != -1 {
		return "SunOS"
	} else if strings.Index(userAgent, "BeOS") != -1 {
		return "BeOS"
	} else if strings.Index(userAgent, "OS/2") != -1 {
		return "OS/2"
	} else if strings.Index(userAgent, "PC") != -1 {
		return "Macintosh"
	} else if strings.Index(userAgent, "AIX") != -1 {
		return "AIX"
	}

	return "未知操作系统"
}

// 匹配设备类型正则
var (
	gXmideaRe = regexp.MustCompile(`xmidea`)
	gPcRe     = regexp.MustCompile(`windows|macintosh|ubuntu|x11`)
	gPadRe    = regexp.MustCompile(`ipad`)
	gMobileRe = regexp.MustCompile(`android|iphone`)
)

// Machine 客户端设备类型判断(pc、mobile、pad).
// Return:
//  - machine: 设备.
func (m *FargoInput) Machine() (machine string) {
	userAgent := m.UserAgent()
	if userAgent == "" {
		return "unknown"
	} else if gXmideaRe.MatchString(userAgent) {
		return "xmidea"
	} else if gPcRe.MatchString(userAgent) {
		return "pc"
	} else if gPadRe.MatchString(userAgent) {
		return "pad"
	} else if gMobileRe.MatchString(userAgent) {
		return "mobile"
	}

	return "pc"
}

// Refer 返回请求的 refer 信息.
// Return:
//  - refer: refer 信息.
func (m *FargoInput) Refer() (refer string) {
	return m.Header("Referer")
}

// SubDomains 返回请求域名的根域名, 例如请求是 www.baidu.com, 那么调用该函数返回 baidu.com.
// Return:
//  - subdomain: 根域名.
func (m *FargoInput) SubDomains() (subdomain string) {
	parts := strings.Split(m.Host(), ".")
	return strings.Join(parts[len(parts)-2:], ".")
}

// Port 返回请求的端口, 例如返回 8080.
// Return:
//  - port: 端口.
func (m *FargoInput) Port() (port int) {
	parts := strings.Split(m.Request.Host, ":")
	if len(parts) == 2 {
		port, _ = strconv.Atoi(parts[1])
		return
	}

	return 80
}

// UserAgent 返回请求的 UserAgent, 例如 Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.57 Safari/537.36.
// Return:
//  - useragent: useragent.
func (m *FargoInput) UserAgent() (useragent string) {
	return m.Header("User-Agent")
}

// Param 在路由设置的时候可以设置参数, 这个是用来获取那些参数的, 例如 Param(":id"), 返回12.
// Return:
//  - Param: 参数.
func (m *FargoInput) Param(key string) (param string) {
	if v, ok := m.Params[key]; ok {
		return v
	}

	return ""
}

// Query 该函数返回 Get 请求和 Post 请求中的所有数据, 和 PHP 中 $_REQUEST 类似.
// Parameters:
//  - key:   key 值
// Return:
//  - value: 值.
func (m *FargoInput) Query(key string) (value string) {
	m.Request.ParseForm()
	return m.Request.Form.Get(key)
}

// PostFormValue return post and put body value
func (m *FargoInput) PostFormValue(key string) (value string) {
	return m.Request.PostFormValue(key)
}

// Header 返回相应的 header 信息, 例如 Header("Accept-Language"), 就返回请求头中对应的信息 zh-CN,zh;q=0.8,en;q=0.6.
// Parameters:
//  - key:    key 值
// Return:
//  - header: header 信息.
func (m *FargoInput) Header(key string) (header string) {
	return m.Request.Header.Get(key)
}

// Cookie 返回请求中的 cookie 数据, 例如 Cookie("username"), 就可以获取请求头中携带的 cookie 信息中 username 对应的值.
// Parameters:
//  - key:    key 值
// Return:
//  - cookie: cookie 信息.
func (m *FargoInput) Cookie(key string) (cookie string) {
	ck, err := m.Request.Cookie(key)
	if err != nil {
		return
	}

	return ck.Value
}

// Session session 是用户可以初始化的信息，默认采用了 Fargo 的 session 模块中的 Session 对象，用来获取存储在服务器端中的数据.
// Parameters:
//  - key:    session key 值
// Return:
//  - cookie: session 信息.
func (m *FargoInput) Session(key interface{}) (val interface{}) {
	return m.CruSession.Get(key)
}

// Body 返回请求 Body 中数据，例如 API 应用中，很多用户直接发送 json 数据包，那么通过 Query 这种函数无法获取数据，就必须通过该函数获取数据.
// Return:
//  - body: body 信息.
func (m *FargoInput) Body() (body []byte) {
	body, _ = ioutil.ReadAll(m.Request.Body)
	m.Request.Body.Close()
	bf := bytes.NewBuffer(body)
	m.Request.Body = ioutil.NopCloser(bf)
	m.RequestBody = body

	return
}

// GetData 获取 context 内容.
// Parameters:
//  - key:  key 值
// Return:
//  - data: context 信息.
func (m *FargoInput) GetData(key interface{}) (data interface{}) {
	if v, ok := m.Data[key]; ok {
		return v
	}

	return
}

// SetData 设置 context 内容.
// Parameters:
//  - key: key 值
//  - val: 设置的值.
func (m *FargoInput) SetData(key, val interface{}) {
	m.Data[key] = val
	return
}

// ParseFormOrMulitForm parseForm or parseMultiForm based on Content-type.
func (m *FargoInput) ParseFormOrMulitForm(maxMemory int64) error {
	if strings.Contains(m.Header("Content-Type"), "multipart/form-data") {
		if err := m.Request.ParseMultipartForm(maxMemory); err != nil {
			return err
		}
	} else if err := m.Request.ParseForm(); err != nil {
		return err
	}

	return nil
}
