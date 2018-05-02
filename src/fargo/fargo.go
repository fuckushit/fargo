package fargo

import (
	"fmt"
	"os"

	"bdlib/comm"
	"bdlib/config"
	"fargo/middleware"
)

// Fargo 框架版本号.
const VERSION = "1.0.0"

// Add 添加路由到 Fargo 应用中.
// Parameters:
// - pattern:        注册的路由 URI, 如 /index, /admin/id 等.
// - c:              controller 的接口对象.
// - mappingMethods: 不定项的路由参数, 用于自定义路由方法时候, 如 “post:postRouter,get:getIndex”.
// Return:
//  - app:           Fargo 对象.
func Add(pattern string, c ControllerInterface, mappingMethods ...string) (app *App) {
	gApp.Router(pattern, c, mappingMethods...)
	return gApp
}

// Run 开跑.
func Run() {

	// 开启session.
	if sessionOn {
		initSession()
	}

	// build 模板文件.
	if err := BuildTemplate(templateDirc); err != nil {
		fmt.Println(comm.WrapError(err))
		os.Exit(1)
	}

	// 中间件初始化.
	middleware.VERSION = VERSION
	middleware.AppName = gAppName
	middleware.RegisterErrorHandler(gExceptionPrefix, g404FilePrefix)

	gApp.Run()
}

// Error 错误处理 写入log.
func Error(err error) {
	nerr := fmt.Errorf("ERROR: %v", err)
	Log.PrintN(3, nerr)
}

// Errorf ...
func Errorf(format string, args ...interface{}) {
	format = "ERROR: " + format
	Log.PrintfN(3, format, args...)
}

// Debug ...
func Debug(err error) {
	if !gDebug {
		return
	}
	nerr := fmt.Errorf("DEBUG: %v", err)
	Log.PrintN(3, nerr)
}

// Debugf ...
func Debugf(format string, args ...interface{}) {
	if !gDebug {
		return
	}
	format = "DEBUG: " + format
	Log.PrintfN(3, format, args...)
}

// Info ...
func Info(err error) {
	nerr := fmt.Errorf("INFO: %v", err)
	Log.PrintN(3, nerr)
}

// Infof ...
func Infof(format string, args ...interface{}) {
	format = "INFO: " + format
	Log.PrintfN(3, format, args...)
}

// ErrorC ...
func ErrorC(err error) {
	if !gDebug {
		return
	}
	Log.ColorLog("[ERRO] %s\n", err)
}

// Configer 获取配置文件操作接口对象.
// Return:
// - cfg: 配置文件接口对象.
func Configer() (cfg config.Configer) {
	return gCfg
}

// GetSection 获取配置文件中的 section.
// Parameters:
// - sectionName: 要获取配置文件中的 section 名, 如配置文件中的 [database], 则此为 database.
// Return:
// - section:    配置文件中一个 section 的内容 map.
// - err:
func GetSection(sectionName string) (section config.Section, err error) {
	return gCfg.GetSection(sectionName)
}

// GetSetting 获取配置文件中某一个 setting 的值, 如 host=10.100.100.100, 则为 10.100.100.100.
// Parameters:
// - sectionName: 要获取配置文件中的 section 名, 如配置文件中的 [database], 则此为 database.
// - keyName:     要获得的设置的名称的名字, 如 host=10.100.100.100 中的 host.
// Return:
// - value:       某一个设置项的值.
// - err:
func GetSetting(sectionName, keyName string) (value string, err error) {
	return gCfg.GetSetting(sectionName, keyName)
}

// GetIntSetting 获取配置文件中一个整型的 setting 值.
// Parameters:
// - sectionName: 要获取配置文件中的 section 名, 如配置文件中的 [database], 则此为 database.
// - keyName:     要获得的设置的名称的名字, 如 host=10.100.100.100 中的 host.
// - dfault:      获取的整型变量不存在的时候的默认值.
// Return:
// - value:       某一个设置项的值.
// - err:
func GetIntSetting(sectionName, keyName string, dfault int64) (value int64, err error) {
	return gCfg.GetIntSetting(sectionName, keyName, dfault)
}

// GetBoolSetting 获取配置文件中一个 bool 的 setting 值.
// Parameters:
// - sectionName: 要获取配置文件中的 section 名, 如配置文件中的 [database], 则此为 database.
// - keyName:     要获得的设置的名称的名字, 如 host=10.100.100.100 中的 host.
// - dfault:      获取的 bool 变量不存在的时候的默认值.
// Return:
// - value:       某一个设置项的值.
// - err:
func GetBoolSetting(sectionName, keyName string, dfault bool) (value bool, err error) {
	return gCfg.GetBoolSetting(sectionName, keyName, dfault)
}

// SetDefaultSection 设置的 web section name, Fargo 框架默认配置文件中的 section 名称为 "web", 这里设置更改.
// Parameters:
// - name: 要设置的 web section name.
func SetDefaultSection(name string) {
	webSection = name
	return
}

// InsertFilter adds a FilterFunc with pattern condition and action constant.
// The pos means action constant including
// fargo.BeforeStatic, fargo.BeforeRouter, fargo.BeforeExec, fargo.AfterExec and fargo.FinishRouter.
// The bool params is for setting the returnOnOutput value (false allows multiple filters to execute)
func InsertFilter(pattern string, pos int, filter FilterFunc, params ...bool) *App {
	gApp.Handlers.InsertFilter(pattern, pos, filter, params...)
	return gApp
}
