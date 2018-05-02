package fargo

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"strings"
	"time"
)

// App defined the app struct.
type App struct {
	Handlers *ControllerRegistor
}

// gApp fargo app
var gApp *App

// NewApp 初始化 fargo 对象.
// Return:
//  - app: fargo 对象.
func NewApp() (app *App) {
	// 初始化路由注册
	cr := NewControllerRegistor()
	app = &App{
		Handlers: cr,
	}
	return
}

// Run fargo 运行
func (a *App) Run() {

	var (
		err         error
		addr, fAddr string
		l           net.Listener
	)

	if httpAddr != "" {
		addr = httpAddr
		fAddr = httpAddr
	} else if httpAddr == "" {
		fAddr = "*"
	}
	if httpPort != 0 {
		addr = fmt.Sprintf("%s:%d", httpAddr, httpPort)
	}
	httpHost = addr

	// 输出框架头部信息
	header := strings.Replace(gHeader, "{{configue}}", *gConfigName, -1)
	header = strings.Replace(header, "{{version}}", VERSION, -1)
	header = strings.Replace(header, "{{host}}", fmt.Sprintf("%s:%d", fAddr, httpPort), -1)
	fmt.Fprintf(os.Stdout, header)

	if err = writePid(); err != nil {
		pwd, _ := os.Getwd()
		Error(fmt.Errorf("%v, pwd:%s, uid:%d", err, pwd, os.Getuid()))
		time.Sleep(100 * time.Microsecond)
		os.Exit(2)
	}

	// panic
	Log.WatchPanic()

	// 封装server - net.http.Server 结构
	// 直接实现了 net.http.Handler 接口，即调用 ServeHTTP 方法.
	if enableSocket {

	} else {
		// 是否使用 fastCgi
		if useFcgi {
			if httpPort == 0 {
				l, err = net.Listen("unix", addr)
			} else {
				l, err = net.Listen("tcp", addr)
			}
			if err != nil {
				Error(err)
				return
			}
			err = fcgi.Serve(l, gApp.Handlers)
		} else {
			if enableHotUpdate {
				// 是否开启热更新
				server := &http.Server{
					Handler:      gApp.Handlers,
					ReadTimeout:  time.Duration(gHTTPServerTimeOut) * time.Second,
					WriteTimeout: time.Duration(gHTTPServerTimeOut) * time.Second,
				}
				laddr, err := net.ResolveTCPAddr("tcp", addr)
				if nil != err {
					Error(err)
					return
				}
				l, err = GetInitListener(laddr)
				theStoppable = newStoppable(l)
				err = server.Serve(theStoppable)
				theStoppable.wg.Wait()
				CloseSelf()
			} else {
				srv := &http.Server{
					Addr:         addr,                                            // 监听的地址和端口
					Handler:      gApp.Handlers,                                   // 所有请求需要调用的Handler
					ReadTimeout:  time.Duration(gHTTPServerTimeOut) * time.Second, // 读的最大Timeout时间
					WriteTimeout: time.Duration(gHTTPServerTimeOut) * time.Second, // 写的最大Timeout时间
				}
				if httpTLS {
					err = srv.ListenAndServeTLS(httpCertFile, httpKeyFile)
				} else {
					err = srv.ListenAndServe()
				}
			}
		}
	}

	if err != nil {
		Error(err)
		time.Sleep(100 * time.Microsecond)
	}
}

// Router 添加 url-pattern 路由规则.
// Parameters:
// - path:           要添加的路由 URI, 如 /index.
// - c:              对应 URI 的 controller 逻辑函数, 实现了 ControllerInterface 接口.
// - mappingMethods: 不定项参数.
// Return:
//  - app: fargo 对象.
func (a *App) Router(path string, c ControllerInterface, mappingMethods ...string) (app *App) {
	gApp.Handlers.Add(path, c, mappingMethods...)
	return gApp
}
