package v1

import (
	"ibanking-scraper/internal/http/handler/healthcheck"
	"ibanking-scraper/internal/http/router"
)

func NewHealthCheckRoute() router.Router {
	r := route.Group("/healthcheck")

	r.GET("/", healthcheck.Fetch())

	return r
}
