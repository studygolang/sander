// Copyright 2018 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://studygolang.com
// Author: meission	meission@aliyun.com

package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"runtime"
	"sync"

	log "github.com/meission/log4go"
)

var (
	logger log.Logger
	bufP   sync.Pool
)

// Init log init
func Init(dir string) {
	if dir != "" {
		logger = log.Logger{}
		log.LogBufferLength = 10240
		// new info writer
		iw := log.NewFileLogWriter(path.Join(dir, "info.log"), false)
		iw.SetRotateDaily(true)
		iw.SetRotateSize(math.MaxInt32)
		iw.SetRotate(true)
		iw.SetFormat("[%D %T] [%L] [%S] %M")
		logger["info"] = &log.Filter{
			Level:     log.INFO,
			LogWriter: iw,
		}
		// new warning writer
		ww := log.NewFileLogWriter(path.Join(dir, "warning.log"), false)
		ww.SetRotateDaily(true)
		ww.SetRotateSize(math.MaxInt32)
		ww.SetRotate(true)
		ww.SetFormat("[%D %T] [%L] [%S] %M")
		logger["warning"] = &log.Filter{
			Level:     log.WARNING,
			LogWriter: ww,
		}
		// new error writer
		ew := log.NewFileLogWriter(path.Join(dir, "error.log"), false)
		ew.SetRotateDaily(true)
		ew.SetRotateSize(math.MaxInt32)
		ew.SetRotate(true)
		ew.SetFormat("[%D %T] [%L] [%S] %M")
		logger["error"] = &log.Filter{
			Level:     log.ERROR,
			LogWriter: ew,
		}

		dw := log.NewFileLogWriter(path.Join(dir, "debug.log"), false)
		dw.SetRotateDaily(true)
		dw.SetRotateSize(math.MaxInt32)
		dw.SetRotate(true)
		dw.SetFormat("[%D %T] [%L] [%S] %M")
		logger["error"] = &log.Filter{
			Level:     log.DEBUG,
			LogWriter: dw,
		}
	}
}

// Close close resource.
func Close() {
	if logger != nil {
		logger.Close()
	}
}

// Info write info log to file or elk.
func Info(format string, args ...interface{}) {
	if logger != nil {
		logger.Info(format, args...)
	}
}

// Warn write warn log to file or elk.
func Warn(format string, args ...interface{}) {
	if logger != nil {
		logger.Warn(format, args...)
	}
}

// Error write error log to file or elk.
func Error(format string, args ...interface{}) {
	if logger != nil {
		logger.Error(format, args...)
	}
}

// Debug write error log to file or elk.
func Debug(format string, args ...interface{}) {
	if logger != nil {
		logger.Debug(format, args...)
	}
}

// InfoTrace write info log to file or elk with traceid.
func InfoTrace(traceID, path, msg string, tm float64) {
	if logger != nil {
		logger.Info("traceid:%s path:%s msg:%s time:%f", traceID, path, msg, tm)
	}
}

// funcName get func name.
func funcName() (fname string) {
	if pc, _, lineno, ok := runtime.Caller(2); ok {
		fname = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
	}
	return
}

// escape escape html characters.
func escape(src string) (dst string) {
	buf, ok := bufP.Get().(*bytes.Buffer)
	if !ok {
		return
	}
	json.HTMLEscape(buf, []byte(src))
	dst = buf.String()
	buf.Reset()
	bufP.Put(buf)
	return
}
