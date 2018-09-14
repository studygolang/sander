package echoutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"sander/db/nosql"
	"sander/logger"

	"github.com/labstack/echo"
	"github.com/polaris1119/goutils"
	"golang.org/x/net/context"
)

// IsAsync 是否异步处理
func IsAsync(ctx echo.Context) bool {
	return goutils.MustBool(ctx.FormValue("async"), false)
}

// WrapContext 返回一个 context.Context 实例
func WrapEchoContext(ctx echo.Context) context.Context {
	r := ctx.Get("request_id")
	return context.WithValue(ctx.Context(), "request_id", r)
}

// WrapContext 返回一个 context.Context 实例。如果 ctx == nil，需要确保 调用 logger.PutLogger()
func WrapContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx
}

func Success(ctx echo.Context, data interface{}) error {
	result := map[string]interface{}{
		"code": 0,
		"msg":  "ok",
		"data": data,
	}

	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	go func(b []byte) {
		if cacheKey := ctx.Get(nosql.CacheKey); cacheKey != nil {
			logger.Debug("cache save:%+v,now:%+v", cacheKey, time.Now())
			nosql.DefaultLRUCache.CompressAndAdd(cacheKey, b, nosql.NewCacheData())
		}
	}(b)

	if ctx.Response().Committed() {
		return nil
	}

	return ctx.JSONBlob(http.StatusOK, b)
}

func Fail(ctx echo.Context, code int, msg string) error {
	if ctx.Response().Committed() {
		return nil
	}

	result := map[string]interface{}{
		"code": code,
		"msg":  msg,
	}

	logger.Error("operate fail:%+v", result)

	return ctx.JSON(http.StatusOK, result)
}

func AsyncResponse(ctx echo.Context, logicInstance interface{}, methodName string, args ...interface{}) error {
	wrapCtx := WrapContext(ctx)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("async response panic:", err)
			}
		}()

		instance := reflect.ValueOf(logicInstance)

		in := make([]reflect.Value, len(args)+1)
		in[0] = reflect.ValueOf(wrapCtx)
		for i, arg := range args {
			in[i+1] = reflect.ValueOf(arg)
		}

		instance.MethodByName(methodName).Call(in)
	}()

	return Success(ctx, nil)
}
