package middleware

import (
	"ibanking-scraper/internal/authorizer"
	"ibanking-scraper/internal/errors"

	"github.com/labstack/echo/v4"
)

func WithAuthorizer(authorizer authorizer.Authorizer, opts ...authorizer.Option) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()

			permit := authorizer.Authorize(ctx, req.Method, c.Path(), opts...)
			if permit.IsPermitted() {
				return next(c)
			}

			return errors.ErrLoginDisabled
		}
	}
}
