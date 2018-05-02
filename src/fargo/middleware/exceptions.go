package middleware

import (
	"fmt"
)

// HTTPException http exceptions
type HTTPException struct {
	// http 状态信息 如 4xx, 5xx
	StatusCode int

	// 描述信息
	Description string
}

// Error 返回 http exceptions 的异常错误信息字符串, 例如 "404 Not Found"
func (h *HTTPException) Error() (err string) {
	return fmt.Sprintf("%d %s", h.StatusCode, h.Description)
}

// http 状态错误集合
// 默认含有 400, 401, 403, 404, 405, 500, 502, 503 和 504
var HTTPExceptionMaps = make(map[int]HTTPException)

func init() {
	// 4xx HTTP Status
	HTTPExceptionMaps[400] = HTTPException{400, "Bad Request"}
	HTTPExceptionMaps[401] = HTTPException{401, "Unauthorized"}
	HTTPExceptionMaps[403] = HTTPException{403, "Forbidden"}
	HTTPExceptionMaps[404] = HTTPException{404, "Not Found"}
	HTTPExceptionMaps[405] = HTTPException{405, "Method Now Allowed"}

	// 5xx HTTP Status
	HTTPExceptionMaps[500] = HTTPException{500, "Internal Server Error"}
	HTTPExceptionMaps[502] = HTTPException{502, "Bad Gateway"}
	HTTPExceptionMaps[503] = HTTPException{503, "Service Unavailable"}
	HTTPExceptionMaps[504] = HTTPException{504, "Gateway Timeout"}
}
