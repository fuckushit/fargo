// Package comm 包含了日志的封装，报警，和一些常用的工具函数
package comm

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

//Logger 错误的包装结构
type Logger struct {
	errChan chan error
}

//Printf 格式化的形式将错误输出到Logger
func (l *Logger) Printf(format string, args ...interface{}) {
	l.errChan <- WrapError(fmt.Errorf(format, args...), int(2))
}

//NewLogger 初始化一个新的Logger对象
func NewLogger(errChan chan error) *Logger {
	return &Logger{errChan}
}

func getFirstInt(args ...interface{}) int {
	count := 0
	var firstArg interface{}
	for _, arg := range args {
		count++
		if count == 1 {
			firstArg = arg
			break
		}
	}
	var trackStack = 1
	if firstArg != nil {
		if _, ok := firstArg.(int); ok {
			trackStack = firstArg.(int)
		}
	}
	return trackStack
}

//WrapError 封装错误信息，获取抛出错误所在的文件名，行号信息。
// WrapError(err) -> caller(1)
// WrapError(err, n) -> caller(n)
func WrapError(err error, args ...interface{}) error {
	trackStack := getFirstInt(args...)
	_, file, line, _ := runtime.Caller(trackStack)
	file = filepath.Base(file)
	errMsg := err.Error()
	if !strings.HasPrefix(errMsg, "DEBUG") && !strings.HasPrefix(errMsg, "INFO") {
		errMsg = "ERROR " + errMsg
	}
	errMsg = fmt.Sprintf("[%s:%d] %s", file, line, errMsg)
	return errors.New(errMsg)
}

// WrapSuccess 封装成功类型，同时获取抛出信息所在的文件名，行号信息
// WrapSuccess(err) -> caller(1)
// WrapSuccess(err, n) -> caller(n)
func WrapSuccess(success string, args ...interface{}) error {
	trackStack := getFirstInt(args...)
	_, file, line, _ := runtime.Caller(trackStack)
	file = filepath.Base(file)
	errMsg := success
	if !strings.HasPrefix(errMsg, "DEBUG") && !strings.HasPrefix(errMsg, "INFO") {
		errMsg = "SUCCESS " + errMsg
	}
	errMsg = fmt.Sprintf("[%s:%d] %s", file, line, errMsg)
	return errors.New(errMsg)
}

// AddrWrapError 封装网络请求错误，加上ip,url信息
// AddrWrapError(err) -> caller(1)
// AddrWrapError(err, n) -> caller(n)
func AddrWrapError(r *http.Request, err error, args ...interface{}) error {
	trackStack := getFirstInt(args...)
	_, file, line, _ := runtime.Caller(trackStack)
	file = filepath.Base(file)
	errMsg := err.Error()
	addr := strings.Split(r.RemoteAddr, ":")[0]
	url := r.URL.Path
	var prefix string
	if !strings.HasPrefix(errMsg, "DEBUG") && !strings.HasPrefix(errMsg, "INFO") {
		prefix = "ERROR"
	}
	errMsg = fmt.Sprintf("[%s:%d] [%s %s] %s %s", file, line, addr, url, prefix, errMsg)
	return errors.New(errMsg)
}

// WrapErrorf 格式化WrapError
func WrapErrorf(format string, args ...interface{}) error {
	return WrapError(fmt.Errorf(format, args...), int(2))
}

// AddrWrapErrorf 格式化AddWrapError
func AddrWrapErrorf(r *http.Request, format string, args ...interface{}) error {
	return AddrWrapError(r, fmt.Errorf(format, args...), 2)
}

//DumpStack 向w输出错误栈信息
func DumpStack(w io.Writer) {
	for i := 1; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		w.Write([]byte(fmt.Sprintf("%s:%d\n", file, line)))
	}
}

//WritePidForWangcai write the pid file.
func WritePidForWangcai() {
	processName := os.Args[0]
	lastIndex := strings.LastIndex(processName, "/")
	processName = processName[lastIndex+1:]

	pid := strconv.Itoa(os.Getpid())
	if pid == "" {
		return
	}
	filename := processName + ".pid"
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer func() {
		file.Close()
	}()
	file.WriteString(pid)

	return
}
