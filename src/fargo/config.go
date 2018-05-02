package fargo

import (
	"bdlib/config"
	"bdlib/logger"
	"errors"
	"fargo/session"
	"flag"
	"html/template"
	"runtime"
)

// 错误配置表
var (
	// ErrMysqlPoolClosed an error for mysql pool has been closed.
	ErrMysqlPoolClosed = errors.New("mysql: get on closed pool")
	// ErrMysqlPoolExhausted an error for connection pool exhausted.
	ErrMysqlPoolExhausted = errors.New("mysql: connection pool exhausted")
	// ErrMysqlNew an error for connect db failed.
	ErrMysqlNew = errors.New("mysql: connect db failed")

	// ErrSessionCfg error for init sesscion config failed.
	ErrSessionCfg = errors.New("session: init sesscion config err")
	// ErrSessionNew error for new session manager failed.
	ErrSessionNew = errors.New("session: new session manager failed")

	// ErrInitWeb an error for get section web failed.
	ErrInitWeb = errors.New("init: get section web failed")
	// ErrInitSPath an error for get section s_path failed.
	ErrInitSPath = errors.New("init: get section s_path failed")
	// ErrInitPath an error for get section path failed.
	ErrInitPath = errors.New("init: get section path failed")
)

// flags.
var (
	gConfigName = flag.String("c", "", "config file name")
	gHost       = flag.String("h", "", "listen host")
)

// 应用设置
var (
	// gAppName 应用名称
	gAppName = "fargo"

	// gServerName response header 中的 server name
	gServerName = "fargoServer"
)

// 通用配置
var (
	gNumCPU = runtime.NumCPU()

	// gTimeFormat time format
	gTimeFormat = "2006-01-02 15:04:05"

	// 模板变量标识
	gTemplateLeft, gTemplateRight string

	// log 日志路径
	gPath, gPrefix string

	// gRandomURL 限制的ip访问用户跳走url
	gRandomURL string

	// gTemplateCache template cache
	gTemplateCache map[string]*template.Template

	// 404以及异常模板路径.
	g404FilePrefix   string
	gExceptionPrefix string
)

// config
var (
	// gCfg 全局配置文件句柄
	gCfg config.Configer

	// webSection 默认设置 section name
	webSection = "web"

	// // gRedisCfg redis 配置
	// gRedisCfg config.Section

	// // gMemCfg memcache 配置
	// gMemCfg config.Section

	// gSRoute 静态文件如 css, js, etc URL
	gSRoute config.Section

	// gHTTPServerTimeOut server 超时时间
	gHTTPServerTimeOut int64

	// maxMemory post 最大内存
	maxMemory int64
)

const (
	// GRedisEncryptKey encrypt key of redis auth.
	GRedisEncryptKey = "2wsxCDE#4rfv"
)

// public config
var (
	// Log ...
	Log *logger.Logger

	// GCfg 配置文件接口
	GCfg config.Configer

	// globalSessions 全局的 session manager.
	globalSessions *session.Manager

	gHostFromCfg string

	// httpAddr addr
	httpAddr string
	// httpPort port
	httpPort int64
	// httpHost Listen Host.
	httpHost string

	// tplPrefix 模板路径前缀
	tplPrefix string
	// templateDirc 模板路径文件
	templateDirc string

	// runMode 网站开发模式, 如 debug 等.
	runMode string
)

// true or false, that is the question
var (
	// gDebug 是否开启debug模式
	gDebug = false

	// // gEnableTemplateFunc 是否使用系统模板函数
	// gEnableTemplateFunc = true

	// // 是否自动加载模板
	// // gEnableAutoRender bool = true

	// autoRender 是否自动渲染模板
	autoRender = true

	// enableStatic 是否开启渲染静态文件, js, css, images, etc...
	enableStatic = true

	// sessionOn 是否开启session
	sessionOn bool

	// useFcgi 是否使用 FastCgi
	useFcgi = false

	// // EnableCmd 是否开启 cmd.
	// EnableCmd = false

	// enableSocket 是否开启 tcp socket.
	enableSocket = false

	// enableHotUpdate 是否开启热更新
	enableHotUpdate = false

	// httpTLS 是否开启 httpTLS
	httpTLS = false
	// httpCertFile ...
	httpCertFile string
	// httpKeyFile ...
	httpKeyFile string

	// enableGzip 是否开启 Gzip
	enableGzip = false

	// enableAccessLog 是否开启 access log.
	enableAccessLog = true

	// directoryIndex flag of display directory index. default is false.
	directoryIndex = false

	// enableXSRF 是否开启 XSRF
	enableXSRF = false
	// XSRFKEY XSRF 加密 key
	XSRFKEY = "fargoxsrf"
	// XSRFExpire xsrf 的生存时间
	XSRFExpire int64

	// EnableLDAPHTTPS LDAP 是否开启 HTTPS
	EnableLDAPHTTPS = false

	// RouterCaseSensitive router case sensitive default is true.
	RouterCaseSensitive = true
)

// 提示信息模板
var (
	gHeader = `~ @Fargo Is A Agile Web Framework, Version: {{version}}
~ Author: chenzhifeng01@baidu.com
~ Parameter:
~	Host  : {{host}}
~ 	Config: {{configue}}
~ Usage:
~	-h=<host:port>
~	-c=<confige file name>
~ `
)
