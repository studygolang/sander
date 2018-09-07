package middleware

import (
	"fmt"
	"net"
	"time"

	"sander/logger"

	"github.com/labstack/echo"
	"github.com/twinj/uuid"
)

const HeaderKey = "X-Request-Id"

// EchoLogger 用于 echo 框架的日志中间件
func EchoLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			start := time.Now()

			req := ctx.Request()
			resp := ctx.Response()

			logger.Info("query params:%+v", ctx.QueryParams())

			remoteAddr := req.RemoteAddress()
			if ip := req.Header().Get(echo.HeaderXRealIP); ip != "" {
				remoteAddr = ip
			} else if ip = req.Header().Get(echo.HeaderXForwardedFor); ip != "" {
				remoteAddr = ip
			} else {
				remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
			}

			id := func(ctx echo.Context) string {
				id := req.Header().Get(HeaderKey)
				if id == "" {
					id = ctx.FormValue("request_id")
					if id == "" {
						id = uuid.NewV4().String()
					}
				}

				ctx.Set("request_id", id)

				return id
			}(ctx)

			resp.Header().Set(HeaderKey, id)

			defer func() {
				method := req.Method()
				path := req.URL().Path()
				if path == "" {
					path = "/"
				}
				size := resp.Size()
				code := resp.Status()

				stop := time.Now()
				// [remoteAddr method path request_id "UA" code time size]
				uri := fmt.Sprintf(`[%s %s %s %s "%s" %d %s %d]`, remoteAddr, method, path, id, req.UserAgent(), code, stop.Sub(start), size)
				logger.Info(uri)
			}()

			if err := next(ctx); err != nil {
				return err
			}
			return nil
		}
	}
}
