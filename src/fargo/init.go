package fargo

import (
	"bdlib/comm"
	"bdlib/config"
	"bdlib/logger"
	"bdlib/util"
	"fargo/session"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
)

// init 新建操作对象
func init() {

	flag.Parse()
	runtime.GOMAXPROCS(gNumCPU / 2)

	if *gConfigName == "" {
		*gConfigName = "etc/web.conf"
	}

	initFargo()

	// 先通过配置文件指定 host, 再通过 flag.
	if strings.TrimSpace(gHostFromCfg) == "" || len(strings.TrimSpace(gHostFromCfg)) == 0 {
		// 监听的端口, 默认为 *:8080.
		if *gHost == "" {
			*gHost = ":8080"
		}
		// 特殊情况处理.
		if *gHost == "localhost" || *gHost == "127.0.0.1" {
			*gHost = "127.0.0.1:8080"
		}

		gHostFromCfg = *gHost
	}

	var err error
	if !strings.HasPrefix(gHostFromCfg, ":") && !strings.Contains(gHostFromCfg, ":") {
		gHostFromCfg = fmt.Sprintf(":%s", gHostFromCfg)
	}
	var port string
	httpAddr, port, err = net.SplitHostPort(gHostFromCfg)
	if err != nil {
		fmt.Println(comm.WrapError(err))
		os.Exit(1)
	}
	httpPort = util.Int64(port)

	// logger 错误日志
	if _, err = os.Stat(gPath); err != nil && os.IsNotExist(err) {
		if err = os.Mkdir(gPath, os.ModePerm); err != nil {
			fmt.Println(comm.WrapError(err))
		}
	}

	Log = logger.NewLogger("fargo")
	logger.DefaultLog = Log
	go Log.WatchErrors(gPrefix, gPath)
	defer Log.Close()

	// 框架执行者
	gApp = NewApp()
}

// 初始化 fargo 全局变量
func initFargo() (err error) {

	// config
	if gCfg, err = config.NewConfiger(*gConfigName); err != nil {
		fmt.Println(comm.WrapError(err))
		os.Exit(1)
	}
	GCfg = gCfg

	// app
	gAppName, _ = gCfg.GetSetting(webSection, "appname")
	if gAppName == "" {
		gAppName = "fargo"
	}

	gServerName, _ = gCfg.GetSetting(webSection, "servername")
	if gServerName == "" {
		gServerName = "fargoServer"
	}

	// 配置文件中 host 信息.
	gHostFromCfg, _ = gCfg.GetSetting(webSection, "host")

	// 404以及异常模板文件路径.
	g404FilePrefix, _ = gCfg.GetSetting(webSection, "404")
	gExceptionPrefix, _ = gCfg.GetSetting(webSection, "exception")

	// log 路径
	gPath, _ = gCfg.GetSetting(webSection, "path")
	gPrefix, _ = gCfg.GetSetting(webSection, "prefix")

	// 是否开启 debug, 默认为 false
	gDebug, _ = gCfg.GetBoolSetting(webSection, "debug", false)

	// 使用模式, 如 api, web, etc..., 默认为 web 应用.
	runMode, _ = gCfg.GetSetting(webSection, "runmode")
	if runMode == "" {
		runMode = "web"
	}

	sessionOn, _ = gCfg.GetBoolSetting(webSection, "sessionOn", false)

	// // 是否加载框架模板函数, 默认为 true
	// gEnableTemplateFunc, _ = gCfg.GetBoolSetting(webSection, "enableTemplteFunc", true)
	// if gEnableTemplateFunc {
	// 	initTemplateFuncs()
	// }

	// 是否自动加载模板, 默认为 true
	autoRender, _ = gCfg.GetBoolSetting(webSection, "autoRender", true)

	// 是否开启渲染静态文件, js, css, images, etc...
	enableStatic, _ = gCfg.GetBoolSetting(webSection, "enablestatic", true)

	// 是否使用 FastCgi, 默认为 false
	useFcgi, _ = gCfg.GetBoolSetting(webSection, "useFcgi", false)

	// // 是否开启 cmd, 默认为 false.
	// EnableCmd, _ = gCfg.GetBoolSetting(webSection, "enableCmd", false)

	// 是否开启 tcp, udp 等 socket, 默认为 false.
	enableSocket, _ = gCfg.GetBoolSetting(webSection, "enableSocket", false)

	// 是否开启热更新, 默认为 false.
	enableHotUpdate, _ = gCfg.GetBoolSetting(webSection, "hotupdate", false)

	// 是否开启 Gzip
	enableGzip, _ = gCfg.GetBoolSetting(webSection, "enableGzip", false)

	// 是否开启 access log.
	enableAccessLog, _ = gCfg.GetBoolSetting(webSection, "enablegaccesslog", true)

	// 是否开启 display directory
	directoryIndex, _ = gCfg.GetBoolSetting(webSection, "directIndex", false)

	// 是否开启 HTTPLTS, 默认为 false
	httpTLS, _ = gCfg.GetBoolSetting(webSection, "httpHTS", false)
	httpCertFile, _ = gCfg.GetSetting(webSection, "httpCertFile")
	httpKeyFile, _ = gCfg.GetSetting(webSection, "httpKeyFile")

	// 是否开启 XSRF
	enableXSRF, _ = gCfg.GetBoolSetting(webSection, "enableXSRF", false)
	XSRFKEY, _ = gCfg.GetSetting(webSection, "xsrfkey")
	if XSRFKEY == "" {
		XSRFKEY = "fargoxsrf"
	}
	XSRFExpire, _ = gCfg.GetIntSetting(webSection, "xsrfExpire", 0)

	// 模板文件路径
	tplPrefix, _ = gCfg.GetSetting(webSection, "tplPrefix")
	templateDirc = tplPrefix

	// 模板变量标识 默认为 {{ }}
	gTemplateLeft, _ = gCfg.GetSetting(webSection, "templateLeft")
	if gTemplateLeft == "" {
		gTemplateLeft = "{{"
	}
	gTemplateRight, _ = gCfg.GetSetting(webSection, "templateRight")
	if gTemplateRight == "" {
		gTemplateRight = "}}"
	}

	// server 超时时间
	gHTTPServerTimeOut, _ = gCfg.GetIntSetting(webSection, "servertimeout", 60)

	// post 最大内存
	maxMemory, _ = gCfg.GetIntSetting(webSection, "maxMemory", 1<<26)

	// 静态文件路径
	if enableStatic {
		gSRoute, err = gCfg.GetSection("s_path")
		if err != nil {
			fmt.Println(comm.WrapError(ErrInitSPath))
			os.Exit(1)
		}
	}

	return
}

// 开启 session
func initSession() {
	var err error
	sessionProvide, _ := gCfg.GetSetting("session", "sessionstore")
	cookieName, _ := gCfg.GetSetting("session", "cookiename")
	sessionMaxLifetime, _ := gCfg.GetIntSetting("session", "sessiontime", 0)
	redisHost, _ := gCfg.GetSetting("session", "redisurl")
	poolsize, _ := gCfg.GetSetting("session", "poolsize")
	auth, _ := gCfg.GetSetting("session", "auth")
	db, _ := gCfg.GetSetting("session", "db")
	timeout, _ := gCfg.GetIntSetting("session", "timeout", 100)

	globalSessions, err = session.NewManager(sessionProvide, cookieName, sessionMaxLifetime, map[interface{}]interface{}{
		"host":     redisHost,
		"db":       db,
		"poolsize": poolsize,
		"auth":     auth,
		"auth_key": GRedisEncryptKey,
		"timeout":  int(timeout),
	})

	if err != nil {
		Error(err)
		fmt.Println(comm.WrapError(err))
		os.Exit(1)
	}

	// TODO GC
	// go globalSessions.GC()

	return
}

// 写进程 pid 文件.
func writePid() (err error) {
	processName := os.Args[0]
	lastIndex := strings.LastIndex(processName, "/")
	processName = processName[lastIndex+1:]
	err = logger.WritePidFile(processName)
	return
}
