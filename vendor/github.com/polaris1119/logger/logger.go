// Copyright 2016 polaris. All rights reserved.
// Use of l source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ERROR = iota
	INFO
	SQL
	DEBUG
)

var levelMap = map[string]int{
	"ERROR": ERROR,
	"INFO":  INFO,
	"SQL":   SQL,
	"DEBUG": DEBUG,
}

var (
	// 日志文件
	debugFile = ""
	sqlFile   = ""
	infoFile  = ""
	errorFile = ""

	accessFile = ""

	level = DEBUG
)

var pool sync.Pool

func init() {
	pool.New = func() interface{} {
		return New(nil)
	}
}

// Init Init("", "INFO")
func Init(logPath, tmpLevel string, prefixs ...string) {
	prefix := ""
	if len(prefixs) > 0 {
		prefix = prefixs[0] + "-"
	}

	os.Mkdir(logPath, 0777)

	debugFile = logPath + "/" + prefix + "debug.log"
	sqlFile = logPath + "/" + prefix + "sql.log"
	infoFile = logPath + "/" + prefix + "info.log"
	errorFile = logPath + "/" + prefix + "error.log"

	accessFile = logPath + "/" + prefix + "access.log"

	level = levelMap[strings.ToUpper(tmpLevel)]
}

func AccessLog(format string, args ...interface{}) {
	file, err := openFile(accessFile)
	if err != nil {
		return
	}
	defer file.Close()
	outputf(file, format, args...)
}

func Infof(format string, args ...interface{}) {
	if level < INFO {
		return
	}

	file, err := openFile(infoFile)
	if err != nil {
		return
	}
	defer file.Close()
	outputf(file, format, args...)
}

func Infoln(args ...interface{}) {
	if level < INFO {
		return
	}

	file, err := openFile(infoFile)
	if err != nil {
		return
	}
	defer file.Close()
	outputln(file, args...)
}

func Errorf(format string, args ...interface{}) {
	file, err := openFile(errorFile)
	if err != nil {
		return
	}
	defer file.Close()
	outputf(file, format, args...)
}

func Errorln(args ...interface{}) {
	file, err := openFile(errorFile)
	if err != nil {
		return
	}
	defer file.Close()
	outputln(file, args...)
}

func Debugf(format string, args ...interface{}) {
	if level < DEBUG {
		return
	}

	file, err := openFile(debugFile)
	if err != nil {
		return
	}
	defer file.Close()
	outputf(file, format, args...)
}

func Debugln(args ...interface{}) {
	if level < DEBUG {
		return
	}

	file, err := openFile(debugFile)
	if err != nil {
		return
	}
	defer file.Close()
	// 加上文件调用和行号
	_, callerFile, line, ok := runtime.Caller(1)
	if ok {
		args = append([]interface{}{"file:", filepath.Base(callerFile), "line:", line}, args...)
	}
	outputln(file, args...)
}

func outputln(file io.Writer, args ...interface{}) {
	_logger := GetLogger()
	_logger.Logger = log.New(file, "", log.Lmicroseconds)
	_logger.Println(args...)
	PutLogger(_logger)
}

func outputf(file io.Writer, format string, args ...interface{}) {
	_logger := GetLogger()
	_logger.Logger = log.New(file, "", log.Lmicroseconds)
	_logger.Printf(format, args...)
	PutLogger(_logger)
}

func openFile(filename string) (*os.File, error) {
	if filename == "" {
		log.Println("[WARNING] You must call logger.Init function First!")
		return nil, fmt.Errorf("[WARNING] You must call logger.Init function First!")
	}

	filename += "-" + time.Now().Format("060102")

	return os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
}

type Logger struct {
	*log.Logger

	// TODO:append 数据时，没有加锁，所以，如果同一个 Logger 实例，多个goroutine并发可能顺序会乱
	debugBuf []interface{}
	sqlBuf   []interface{}
	infoBuf  []interface{}
	errorBuf []interface{}

	ctx context.Context
}

// GetLogger returns `*Logger` from the sync.Pool. You must return the *Logger by
// calling `PutLogger()`.
func GetLogger() *Logger {
	return pool.Get().(*Logger)
}

// PutLogger returns `*Logger` instance back to the sync.Pool. You must call it after
// `GetLogger()`.
func PutLogger(_logger *Logger) {
	pool.Put(_logger)
}

func New(out io.Writer) *Logger {
	if out == nil {
		return &Logger{}
	}

	return &Logger{
		Logger: log.New(out, "", log.Lmicroseconds),
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.appendInfo(fmt.Sprintf(format, args...))
}

func (l *Logger) Infoln(args ...interface{}) {
	l.appendInfo(fmt.Sprintln(args...))
}

func (l *Logger) appendInfo(info string) {
	if level < INFO {
		return
	}

	if len(l.infoBuf) == 0 {
		l.resetBuf()
	}

	l.infoBuf = append(l.infoBuf, info)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.appendError(fmt.Sprintf(format, args...))
}

func (l *Logger) Errorln(args ...interface{}) {
	l.appendError(fmt.Sprintln(args...))
}

func (l *Logger) appendError(errstr string) {
	if len(l.infoBuf) == 0 {
		l.resetBuf()
	}
	l.errorBuf = append(l.errorBuf, errstr)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.appendDebug(fmt.Sprintf(format, args...))
}

func (l *Logger) Debugln(args ...interface{}) {
	l.appendDebug(fmt.Sprintln(args...))
}

func (l *Logger) appendDebug(debugstr string) {
	if level < DEBUG {
		return
	}

	if len(l.infoBuf) == 0 {
		l.resetBuf()
	}

	l.debugBuf = append(l.debugBuf, "["+strconv.FormatInt(time.Now().Unix(), 10)+"]"+debugstr)
}

func (l *Logger) Sqlf(format string, args ...interface{}) {
	l.appendSql(fmt.Sprintf(format, args...))
}

func (l *Logger) Sqlln(args ...interface{}) {
	l.appendSql(fmt.Sprintln(args...))
}

func (l *Logger) appendSql(info string) {
	if level < SQL {
		return
	}

	if len(l.infoBuf) == 0 {
		l.resetBuf()
	}

	l.sqlBuf = append(l.sqlBuf, info)
}

func (l *Logger) SetContext(ctx context.Context) {
	l.ctx = ctx
}

func (l *Logger) resetBuf() {
	l.debugBuf = make([]interface{}, 1, 20)
	l.sqlBuf = make([]interface{}, 1, 20)
	l.infoBuf = make([]interface{}, 1, 20)
	l.errorBuf = make([]interface{}, 1, 20)
}

func (l *Logger) Flush() {

	var (
		file *os.File
		err  error

		uri interface{} = ""
	)

	if l.ctx != nil {
		uri = l.ctx.Value("uri")
	}

	if len(l.debugBuf) > 1 {
		file, err = openFile(debugFile)
		if err == nil {
			l.Logger = log.New(file, "", log.Lmicroseconds)
			defer file.Close()

			l.debugBuf[0] = uri
			l.Println(l.debugBuf...)
		}
	}

	if len(l.sqlBuf) > 1 {
		file, err = openFile(sqlFile)
		if err == nil {
			l.Logger = log.New(file, "", log.Lmicroseconds)
			defer file.Close()

			if uri == "" {
				l.sqlBuf[0] = "[SQL]"
			} else {
				l.sqlBuf[0] = uri
			}

			l.Println(l.sqlBuf...)
		}
	}

	if len(l.infoBuf) > 1 {
		file, err = openFile(infoFile)
		if err == nil {
			l.Logger = log.New(file, "", log.Lmicroseconds)
			defer file.Close()

			l.infoBuf[0] = uri
			l.Println(l.infoBuf...)
		}
	}

	if len(l.errorBuf) > 1 {
		file, err = openFile(errorFile)
		if err == nil {
			l.Logger = log.New(file, "", log.Lmicroseconds)
			defer file.Close()

			l.errorBuf[0] = uri
			l.Println(l.errorBuf...)
		}
	}

	l.resetBuf()
}
