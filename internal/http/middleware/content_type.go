package middleware

import (
	"ibanking-scraper/internal/errors"

	"github.com/labstack/echo/v4"
)

func WithContentType(contentType string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			content := c.Request().Header.Get(echo.HeaderContentType)
			if content != contentType {
				return errors.ErrWrongContentType
			}
			return next(c)
		}
	}
}
