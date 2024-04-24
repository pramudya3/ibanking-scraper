package http

import (
	v1 "ibanking-scraper/internal/http/router/v1"
	"ibanking-scraper/internal/http/server"

	"go.uber.org/fx"
)

var Module = fx.Options(
	v1.Module,
	server.Module,
)
