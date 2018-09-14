package middleware

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo"
)

// AuthConfig .
type AuthConfig struct {
	signature func(url.Values, string) string
	secretKey string
}

// DefaultAuthConfig .
var DefaultAuthConfig = &AuthConfig{}

// EchoAuth .
func EchoAuth() echo.MiddlewareFunc {
	return EchoAuthWithConfig(DefaultAuthConfig)
}

// EchoAuthWithConfig 用于 echo 框架的签名校验中间件
func EchoAuthWithConfig(authConfig *AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			sign := authConfig.signature(ctx.FormParams(), authConfig.secretKey)
			if sign != ctx.FormValue("sign") {
				return ctx.String(http.StatusBadRequest, `400 Bad Request`)
			}
			return next(ctx)
		}
	}
}
