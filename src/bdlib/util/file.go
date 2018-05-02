package util

import (
	"bdlib/logger"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func FileExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func ReadFile(filename string) (rs string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		logger.Error(err)
		return
	}
	defer f.Close()

	var fd []byte
	fd, err = ioutil.ReadAll(f)
	if err != nil {
		logger.Error(err)
		return
	}
	rs = string(fd)
	return
}

func ReadLine(filename string) (rs []string, err error) {
	f, err := os.Open(filename)
	if err != nil {
		logger.Error(err)
		return
	}
	defer f.Close()

	line := ""
	buff := bufio.NewReader(f)
	for {
		line, err = buff.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil || io.EOF == err {
			if io.EOF == err {
				if line != "" {
					rs = append(rs, line)
				}
				err = nil
			} else {
				logger.Error(err)
			}
			break
		}
		rs = append(rs, line)
	}
	return
}

func WriteLine(filename string, data []interface{}) (err error) {
	var f *os.File
	if FileExist(filename) {
		if err = os.Remove(filename); err != nil {
			logger.Error(err)
			return
		}
	}
	if f, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err != nil {
		logger.Error(err)
		return
	}
	defer f.Close()
	for _, s := range data {
		line := ToString(s) + "\n"
		if _, err = io.WriteString(f, line); err != nil {
			logger.Error(err)
			return
		}
	}
	return
}

// SelfPath 获取当前路径.
func SelfPath() (path string) {
	path, _ = filepath.Abs(os.Args[0])

	return
}

// SelfDir 获取当前文件夹.
func SelfDir() (dir string) {
	dir = filepath.Dir(SelfPath())

	return
}

// FileExists 返回文件或者目录是否存在
func FileExists(name string) (is bool) {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}

// SearchFile 在 paths 下搜索文件
// 常用于在 /etc ~/ 下面搜索配置文件
func SearchFile(filename string, paths ...string) (fullpath string, err error) {
	for _, path := range paths {
		if fullpath = filepath.Join(path, filename); FileExists(fullpath) {
			return
		}
	}
	err = fmt.Errorf("%s not fount in paths", filename)

	return
}

// GrepFile grep 文件内容
// 使用方法: GrepFile(`^hello`, "hello.txt")
// \n 在读的时候会被忽略
func GrepFile(patten string, filename string) (lines []string, err error) {
	re, err := regexp.Compile(patten)
	if err != nil {
		return
	}

	fd, err := os.Open(filename)
	if err != nil {
		return
	}
	lines = make([]string, 0)
	reader := bufio.NewReader(fd)
	prefix := ""
	for {
		byteLine, isPrefix, er := reader.ReadLine()
		if er != nil && er != io.EOF {
			return
		}
		line := string(byteLine)
		if isPrefix {
			prefix += line
			continue
		}

		line = prefix + line
		if re.MatchString(line) {
			lines = append(lines, line)
		}
		if er == io.EOF {
			break
		}
	}

	return
}
