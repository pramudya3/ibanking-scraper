package v1

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/authorizer"
	"ibanking-scraper/internal/http/handler/log"
	"ibanking-scraper/internal/http/router"
)

func NewLogRoute(c domain.LogUsecase) router.Router {
	r := route.Group("/internet-banking/logs")

	r.GET("/:id/", log.GetLogByAkun(c)).Restricted(authorizer.ForAuthenticatedUser())

	return r
}
