package server

import (
	"ibanking-scraper/config"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/authorizer"
	"ibanking-scraper/internal/http/middleware"
	"ibanking-scraper/internal/http/router"
	"ibanking-scraper/internal/logger"
	"ibanking-scraper/internal/validator"

	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

var Module = fx.Provide(New)

type EchoServerParams struct {
	fx.In

	AppConfig  *config.Config
	AppLogger  *logger.Logger
	Routers    router.Routers `group:"router"`
	JWTDecoder domain.JWTDecoder
	IAMUsecase domain.IAMUsecase
}

// New create an instance of Echo server
func New(p EchoServerParams) *echo.Echo {
	log := p.AppLogger.Logger()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger = p.AppLogger
	e.Validator = validator.New()
	e.Binder = NewBinder()
	e.HTTPErrorHandler = ErrorHandler

	e.Pre(middleware.WithTrailingSlash())
	e.Use(middleware.WithCORS())
	e.Use(middleware.WithAccessLogger(log))
	e.Use(middleware.WithSecure())
	e.Use(middleware.WithJWTDecoder(p.AppConfig, p.JWTDecoder))

	p.Routers.Mounts(e, authorizer.New(p.IAMUsecase))

	return e
}
