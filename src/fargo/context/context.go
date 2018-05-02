package context

import (
	"net/http"
)

// Context 上下文结构
// FargoInput 和 FargoOutput 提供了一系列封装便于对 request 和 response 进行操作
type Context struct {
	Input          *FargoInput
	Output         *FargoOutput
	Request        *http.Request
	ResponseWriter http.ResponseWriter
}

// Redirect 带有 http header status code 的强制跳转.
// Parameters:
// - status:   跳转的状态码, 如 301, 302 等.
// - localurl: 跳转的 url.
func (c *Context) Redirect(status int, localurl string) {
	c.Output.Header("Location", localurl)
	c.Output.SetStatus(status)
}

// WriteString 输出一个字符串到 response body.
// Parameters:
// - content: 要输出到 body 的内容.
func (c *Context) WriteString(content string) {
	c.Output.Body([]byte(content))
}

// WriteByte 输出一个 []byte 到 response body.
// Parameters:
// - content: 要输出到 body 的内容.
func (c *Context) WriteByte(content []byte) {
	c.Output.Body(content)
}

// GetCookie 获取 cookie.
// Parameters:
// - key:     要获取的 cookie 的 key.
// Return:
//  - cookie: 要获取的 cookie 值.
func (c *Context) GetCookie(key string) (cookie string) {
	return c.Input.Cookie(key)
}

// SetCookie 设置 cookie.
// Parameters:
// - name:   要设置的 cookie 的 key.
// - value:  要设置的 cookie 的 值.
// - others: 其他 cookie 的选项, 如 path、HttpOnly 等.
func (c *Context) SetCookie(name string, value string, others ...interface{}) {
	c.Output.Cookie(name, value, others...)
}
