package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// WithTrailingSlash will adds a trailing slash to the request URI
func WithTrailingSlash() echo.MiddlewareFunc {
	// return middleware.AddTrailingSlashWithConfig(
	// 	middleware.TrailingSlashConfig{
	// 		RedirectCode: http.StatusMovedPermanently,
	// 	},
	// )
	return middleware.AddTrailingSlash()
}
