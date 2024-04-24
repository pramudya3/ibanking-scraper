package middleware

import (
	"context"
	"ibanking-scraper/config"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/pkg/constant"
	"strings"

	"github.com/labstack/echo/v4"
)

func WithJWTDecoder(config *config.Config, decoder domain.JWTDecoder) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				return next(setAnonymousUser(c))
			}

			var authBearerKey = "Bearer"
			token := strings.Split(authHeader, " ")
			if len(token) != 2 || (len(token) == 2 && token[0] != authBearerKey) {
				if token[0] == "Basic" {
					return next(c)
				}
				return errors.ErrUnauthorized
			}

			user, err := decoder.Decode(ctx, token[1], config.SecretAccessJWT)
			if err != nil {
				return err
			}

			reqCtx := context.WithValue(ctx, constant.ContextKeyUser, user)
			req := c.Request().Clone(reqCtx)
			c.SetRequest(req)
			return next(c)
		}
	}
}

func setAnonymousUser(c echo.Context) echo.Context {
	ctx := c.Request().Context()
	reqCtx := context.WithValue(ctx, constant.ContextKeyUser, domain.AnonymousUser)
	req := c.Request().Clone(reqCtx)
	c.SetRequest(req)
	return c
}
