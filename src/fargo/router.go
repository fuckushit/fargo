package fargo

import (
	"bdlib/comm"
	"bdlib/util"
	"bufio"
	fargocontext "fargo/context"
	"fargo/middleware"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// BEFORE_STATIC default filter execution points
	BEFORE_STATIC = iota
	// BEFORE_ROUTER _
	BEFORE_ROUTER
	// BEFORE_EXEC _
	BEFORE_EXEC
	// AFTER_EXEC _
	AFTER_EXEC
	// FINISH_ROUTER _
	FINISH_ROUTER
)

var (
	// HTTPMETHOD 支持的 http 方法.
	HTTPMETHOD = []string{"get", "post", "put", "delete", "patch", "options", "head"}

	// exceptMethod fargo.Controller 支持的方法 但是不会反射到 AutoRouter 上.
	exceptMethod = []string{"Init", "Prepare", "Finish", "Render", "RenderString",
		"RenderBytes", "Redirect", "Input", "ParseForm", "GetString", "GetStrings", "GetInt", "GetBool",
		"GetFloat", "GetFile", "SaveToFile", "StartSession", "SetSession", "GetSession",
		"DelSession", "SessionRegenerate", "DestroySession", "IsAjax", "XsrfToken", "CheckXSRFCookie"}
)

// controllerInfo 每一个 controller router 的信息集合.
type controllerInfo struct {
	// 注册的 URI.
	pattern string

	// 注册的正则路由的编译对象.
	regex *regexp.Regexp

	// 参数集.
	params map[int]string

	// controller 对应的类型, 通过反射获得.
	controllerType reflect.Type

	// controller 对应的方法集合.
	methods map[string]string

	// 判断 controller 是否含有方法.
	hasMethod bool
}

// ControllerRegistor controller router 注册, 包含路由规则(正则 + 固定路由), 以及 controller handler,
// 在添加路由的时候就将路由信息加入到了这里.
type ControllerRegistor struct {
	// 正则路由
	routers []*controllerInfo

	// 固定路由
	fixrouters []*controllerInfo

	// 是否开启过滤
	enableFilter bool

	// 是否开启智能路由
	enableAuto bool

	// 过滤路由.
	filters map[int][]*FilterRouter

	// 智能路由 key: controller key: method value: reflect.type
	autoRouter map[string]map[string]reflect.Type
}

// NewControllerRegistor 初始化新建一个路由集合.
func NewControllerRegistor() (ct *ControllerRegistor) {
	return &ControllerRegistor{
		routers:    make([]*controllerInfo, 0),
		autoRouter: make(map[string]map[string]reflect.Type),
		filters:    make(map[int][]*FilterRouter),
	}
}

// Add 添加路由的 handler 和 URI 到路由集合对象中,
// 使用方法为:
// - 默认的模式是一个 URI 对应的 一个方法的规则, 如 Add("/user", &UserController{}),
// - 还可以使用自定义方法的路由规则, 如 Add("/api/list", &RestController{}, "*:ListFood"),
// - POST 请求到指定方法: Add("/api/create", &RestController{}, "post:CreateFood"),
// - PUT 请求到指定方法: Add("/api/update", &RestController{}, "put:UpdateFood"),
// - DELETE 请求到指定方法: Add("/api/delete", &RestController{}, "delete:DeleteFood"),
// - 同时多个请求到指定方法: Add("/api", &RestController{}, "get,post:ApiFunc"),
// - 同时指定多种不同对应关系: Add("/admin", &AdminController{}, "get:GetFunc;post:PostFunc").
// Parameters:
// - pattern:        注册的路由 URI, 如 /index, /admin/id 等.
// - c:              controller 的接口对象.
// - mappingMethods: 不定项的路由参数, 用于自定义路由方法时候, 如 “post:postRouter,get:getIndex”.
func (p *ControllerRegistor) Add(pattern string, c ControllerInterface, mappingMethods ...string) {
	j := 0
	params := make(map[int]string)
	parts := strings.Split(pattern, "/")
	for i, part := range parts {
		// 正则路由.
		if strings.HasPrefix(part, ":") {
			expr := "(.*)"
			if index := strings.Index(part, "("); index != -1 {
				expr = part[index:]
				part = part[:index]
				// 匹配 /user/:id:int 正则路由到 /user/:id:([0-9]+).
				// 匹配 /post/:username:string 正则路由到 /user/:username:([\w]+).
			} else if lindex := strings.LastIndex(part, ":"); lindex != 0 {
				switch part[lindex:] {
				case ":int":
					expr = "([0-9]+)"
					part = part[:lindex]
				case ":string":
					expr = `([\w]+)`
					part = part[:lindex]
				}
			}
			params[j] = part
			parts[i] = expr
			j++
		}
		// *全匹配方式 (*.*) 通过 :path 和 :ext 获取参数.
		// (.*) 通过 :splat 获取参数.
		if strings.HasPrefix(part, "*") {
			expr := "(.*)"
			if part == "*.*" {
				params[j] = ":path"
				parts[i] = "([^.]+).([^.]+)"
				j++
				params[j] = ":ext"
				j++
			} else {
				params[j] = ":splat"
				parts[i] = expr
				j++
			}
		}
		// 匹配 someprefix:id(xxx).html 的 url.
		if strings.Contains(part, ":") && strings.Contains(part, "(") && strings.Contains(part, ")") {
			var out []rune
			var start bool
			var startexp bool
			var param []rune
			var expt []rune
			for _, v := range part {
				if start {
					if v != '(' {
						param = append(param, v)
						continue
					}
				}
				if startexp {
					if v != ')' {
						expt = append(expt, v)
						continue
					}
				}
				if v == ':' {
					param = make([]rune, 0)
					param = append(param, ':')
					start = true
				} else if v == '(' {
					startexp = true
					start = false
					params[j] = string(param)
					j++
					expt = make([]rune, 0)
					expt = append(expt, '(')
				} else if v == ')' {
					startexp = false
					expt = append(expt, ')')
					out = append(out, expt...)
				} else {
					out = append(out, v)
				}
			}
			parts[i] = string(out)
		}
	}

	// TODO 智能路由 AutoRoute 模式, 通过反射获取集合.

	reflectVal := reflect.ValueOf(c)
	t := reflect.Indirect(reflectVal).Type()
	methods := make(map[string]string)

	// 创建固定路由.
	if j == 0 {
		route := &controllerInfo{}
		route.pattern = pattern
		route.controllerType = t
		route.methods = methods
		if len(methods) > 0 {
			route.hasMethod = true
		}
		p.fixrouters = append(p.fixrouters, route)
	} else {
		// 创建正则路由.
		// 重新更新 url 模式为正则表达式并且编译正则表达式.
		pattern = strings.Join(parts, "/")
		regex, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Println(comm.WrapError(err))
			return
		}

		// 创建路由.
		route := &controllerInfo{}
		route.regex = regex
		route.params = params
		route.pattern = pattern
		route.methods = methods
		if len(methods) > 0 {
			route.hasMethod = true
		}
		route.controllerType = t
		p.routers = append(p.routers, route)
	}
}

// getErrorHandler 从 middleware 获取 err handler
func (p *ControllerRegistor) getErrorHandler(errorCode string) (handler func(rw http.ResponseWriter, r *http.Request)) {
	handler = middleware.SimpleServerError
	var ok = true
	if errorCode != "" {
		handler, ok = middleware.ErrorMaps[errorCode]
		if !ok {
			handler, ok = middleware.ErrorMaps["500"]
		}
		if !ok || handler == nil {
			handler = middleware.SimpleServerError
		}
	}

	return
}

// getRunMethod 从 request header 或者表单中获取请求的方法名称,
// 有些时候某些浏览器不能创建 create 和 delete 请求, 使用 _method 代替.
// Parameters:
// - method:  http 方法, 如 GET、POST 等.
// - context: fargo 上下文.
// - router:  每一个对应的 controller 集合.
// Return；
// - m:       运行的 http 方法.
func (p *ControllerRegistor) getRunMethod(method string, context *fargocontext.Context, router *controllerInfo) (m string) {
	method = strings.ToLower(method)
	if method == "post" && strings.ToLower(context.Input.Query("_method")) == "put" {
		method = "put"
	}
	if method == "post" && strings.ToLower(context.Input.Query("_method")) == "delete" {
		method = "delete"
	}
	if router.hasMethod {
		if m, ok := router.methods[method]; ok {
			return m
		} else if m, ok = router.methods["*"]; ok {
			return m
		} else {
			return ""
		}
	} else {
		return strings.Title(method)
	}
}

// responseWriter 是 http.ResponseWriter 的封装,
// started 设置为 true 的时候表示 此 ResponseWriter 不会被其他 handler 执行.
type responseWriter struct {
	// http 输出.
	writer http.ResponseWriter

	// 是否已经开始.
	started bool

	// response 的 状态.
	status int

	// 输出编码状态, 是否需要压缩等.
	contentEncoding string
}

// Header 返回 发送到 WriteHeader 的 header map.
// Return；
// - h: http response header.
func (r *responseWriter) Header() (h http.Header) {
	return r.writer.Header()
}

// InitHeadContent 初始化 content-length header, 是否需要 gzip, 长度 等.
// Parameters:
// - contentLength: content 长度.
func (r *responseWriter) InitHeadContent(contentLength int64) {
	if "gzip" == r.contentEncoding {
		r.Header().Set("Content-Encoding", "gzip")
	} else if "deflate" == r.contentEncoding {
		r.Header().Set("Content-Encoding", "deflate")
	} else {
		r.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	}
}

// Write 方法将数据作为 HTTP 的回应写入连接.
// 并且设置 started 置为 true. started 为 true 意味这回应已经被发送.
// Parameters:
// - p: 写回的数据.
func (r *responseWriter) Write(p []byte) (int, error) {
	r.started = true
	return r.writer.Write(p)
}

// WriteHeader 发送带有 status code 的 HTTP response header,
// 并且设置 started 置为 true.
// Parameters:
// - code: 状态码.
func (r *responseWriter) WriteHeader(code int) {
	r.started = true
	r.status = code
	r.writer.WriteHeader(code)
}

// Hijack 将 writer 转换成 hijack.
// HTTP 包中封装了 Hijacker 接口, 允许程序被接替, 详见 net/http 包.
func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := r.writer.(http.Hijacker)
	if !ok {
		println("supported?")
		return nil, nil, fmt.Errorf("webserver doesn't support hijacking")
	}
	return hj.Hijack()
}

// CloseNotify 方法用于获取客户端连接是否断开.
// 返回是个channel，如果从channel中读取到数据，说明连接断开了
func (r *responseWriter) CloseNotify() <-chan bool {
	if cnotifier, ok := r.writer.(http.CloseNotifier); ok {
		// http.ResponseWriter 实现了CloseNotify接口
		return cnotifier.CloseNotify()
	}
	// 一般不会到下面,下面的channel里不会有数据
	return make(<-chan bool, 1)
}

// InsertFilter Add a FilterFunc with pattern rule and action constant.
// The bool params is for setting the returnOnOutput value (false allows multiple filters to execute)
func (p *ControllerRegistor) InsertFilter(pattern string, pos int, filter FilterFunc, params ...bool) error {

	mr := new(FilterRouter)
	mr.tree = NewTree()
	mr.pattern = pattern
	mr.filterFunc = filter
	if !RouterCaseSensitive {
		pattern = strings.ToLower(pattern)
	}
	if len(params) == 0 {
		mr.returnOnOutput = true
	} else {
		mr.returnOnOutput = params[0]
	}
	mr.tree.AddRouter(pattern, true)
	return p.insertFilterRouter(pos, mr)
}

// add Filter into
func (p *ControllerRegistor) insertFilterRouter(pos int, mr *FilterRouter) error {
	p.filters[pos] = append(p.filters[pos], mr)
	p.enableFilter = true
	return nil
}

// ServeHTTP ControllerRegistor 实现了 http.Handler 接口, 而此接口实现了 ServeHTTP 方法,
// 意味着每次 server accept 请求则会执行此方法,
// 将请求和路由集合进行匹配, 通过反射进行路由.
func (p *ControllerRegistor) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			Log.Printf("the request url is %s ", r.URL.Path)
			Log.Printf("crashed error is %v ", err)
			Log.DumpStack()
			if _, ok := err.(middleware.HTTPException); ok {
				// 4xx 和 5xx 错误.
			} else {
				handler := p.getErrorHandler(fmt.Sprint(err))
				handler(rw, r)
				return
			}
		}
	}()

	var (
		findrouter bool
		runMethod  string
		runrouter  reflect.Type
	)

	// 请求开始时间.
	requestPath := r.URL.Path
	beforeRequestTime := time.Now()
	requestUnix := beforeRequestTime.Unix()

	params := make(map[string]string)
	w := &responseWriter{writer: rw}
	w.Header().Set("Server", gServerName)

	// 初始化 context 将 response 和 request 包入 Context 中
	context := &fargocontext.Context{
		ResponseWriter: w,
		Request:        r,
		Input:          fargocontext.NewInput(r),
		Output:         fargocontext.NewOutput(),
	}
	context.Output.Context = context
	context.Output.EnableGzip = enableGzip

	var urlPath string
	if !RouterCaseSensitive {
		urlPath = strings.ToLower(r.URL.Path)
	} else {
		urlPath = r.URL.Path
	}

	// defined filter function
	doFilter := func(pos int) (started bool) {
		if p.enableFilter {
			if l, ok := p.filters[pos]; ok {
				for _, filterR := range l {
					if ok, p := filterR.ValidRouter(urlPath); ok {
						context.Input.Params = p
						filterR.filterFunc(context)
						if filterR.returnOnOutput && w.started {
							return true
						}
					}
				}
			}
		}

		return false
	}

	if context.Input.IsWebsocket() {
		context.ResponseWriter = rw
	}

	// session init.
	if sessionOn && runMode == "web" {
		context.Input.CruSession = globalSessions.SessionStart(w, r)
	}

	// 检测方法.
	if !util.InSlice(strings.ToLower(r.Method), HTTPMETHOD) {
		http.Error(w, "Method Not Allowed", 405)
		return
	}

	if r.Method != "GET" && r.Method != "HEAD" {
		context.Input.Body()
		context.Input.ParseFormOrMulitForm(maxMemory)
	}

	// static file 前的过滤函数.
	if doFilter(BEFORE_STATIC) {
		return
	}

	// 静态文件.
	if runMode == "web" {
		for prefix, staticDir := range gSRoute {
			if r.URL.Path == "/favicon.ico" {
				file := staticDir + r.URL.Path
				http.ServeFile(w, r, file)
				w.started = true
			}
			if strings.HasPrefix(r.URL.Path, prefix) {
				file := staticDir + r.URL.Path[len(prefix):]
				finfo, err := os.Stat(file)
				if err != nil {
					http.NotFound(w, r)
					continue
				}
				// 如果访问的是文件夹并且设置 directoryIndex 为 false.
				if finfo.IsDir() && !directoryIndex {
					middleware.Exception("403", rw, r, "403 Forbidden")
					continue
				}

				http.ServeFile(w, r, file)
				w.started = true
			}
		}
	}

	if doFilter(BEFORE_ROUTER) {
		return
	}

	// 查找固定路由的路径.
	for _, route := range p.fixrouters {
		n := len(requestPath)
		if requestPath == route.pattern {
			runMethod = p.getRunMethod(r.Method, context, route)
			if runMethod != "" {
				runrouter = route.controllerType
				findrouter = true
				break
			}
		}
		// 保证模式 /admin  下 url /admin 200 /admin/ 200.
		// 保证模式 /admin/ 下 url /admin 301 /admin/ 200.
		if requestPath[n-1] != '/' && requestPath+"/" == route.pattern {
			http.Redirect(w, r, requestPath+"/", 301)
			continue
		}
		if requestPath[n-1] == '/' && route.pattern+"/" == requestPath {
			runMethod = p.getRunMethod(r.Method, context, route)
			if runMethod != "" {
				runrouter = route.controllerType
				findrouter = true
				break
			}
		}
	}

	// 查找正则路由.
	if !findrouter {
		for _, route := range p.routers {

			// 检测正则是否匹配 url.
			if !route.regex.MatchString(requestPath) {
				continue
			}

			// 双重检测正则是否和 url 模式匹配.
			matches := route.regex.FindStringSubmatch(requestPath)
			if len(matches[0]) != len(requestPath) {
				continue
			}

			if len(route.params) > 0 {
				// 在 query 参数 map 中添加 url 参数.
				values := r.URL.Query()
				for i, match := range matches[1:] {
					values.Add(route.params[i], match)
					params[route.params[i]] = match
				}
				// 重组.
				r.URL.RawQuery = url.Values(values).Encode()
			}
			runMethod = p.getRunMethod(r.Method, context, route)
			if runMethod != "" {
				runrouter = route.controllerType
				context.Input.Params = params
				findrouter = true
				break
			}
		}
	}

	// 如果路由还没有找到, 抛出 404 页面.
	if !findrouter {
		middleware.Exception("404", rw, r, "")
		return
	}

	// 找到了路由则转向对应的控制层方法上.
	if findrouter {
		// execute 前的 filter.
		if doFilter(BEFORE_EXEC) {
			return
		}

		// 调用 handler.
		c := reflect.New(runrouter)
		execController, ok := c.Interface().(ControllerInterface)
		if !ok {
			Log.Print(fmt.Errorf("controller is not ControllerInterface"))
			return
		}

		// 执行 controller.Init() 方法, 进行 controller 初始化.
		execController.Init(context, runrouter.Name(), runMethod, c.Interface())

		// 如果设置了 XSRF, 则 检测 cookie 中 是否有任何 _csrf
		if enableXSRF {
			execController.XsrfToken()
			if r.Method == "POST" || r.Method == "DELETE" || r.Method == "PUT" || (r.Method == "POST" && (r.Form.Get("_method") == "put")) {
				execController.CheckXSRFCookie()
			}
		}

		// 执行 prepare funtion
		execController.Prepare()

		// 执行 filter 函数
		if !execController.Filter() {
			afterRequestTime := time.Now()
			requestTime := afterRequestTime.Sub(beforeRequestTime)
			execController.accessLog(requestTime, requestUnix)
			execController.Finish()
			return
		}

		// 执行主体
		if !w.started {
			switch runMethod {
			case "Get":
				execController.Get()
			case "Post":
				execController.Post()
			case "Delete":
				execController.Delete()
			case "Put":
				execController.Put()
			case "Head":
				execController.Head()
			case "Patch":
				execController.Patch()
			case "Options":
				execController.Options()
			default:
				var in = make([]reflect.Value, 0)
				method := c.MethodByName(runMethod)
				method.Call(in)
			}

			// 请求使用时间以及当前请求时间戳, 并记录 access log.
			if enableAccessLog {
				afterRequestTime := time.Now()
				requestTime := afterRequestTime.Sub(beforeRequestTime)
				execController.accessLog(requestTime, requestUnix)
			}

			// 渲染模板
			if !w.started && !context.Input.IsWebsocket() {
				if autoRender {
					if err := execController.Render(); err != nil {
						Error(err)
						return
					}
				}
			}
		}

		// 完成，释放资源
		execController.Finish()

		// execute 之后的 filter.
		if doFilter(AFTER_EXEC) {
			return
		}
	}
}
