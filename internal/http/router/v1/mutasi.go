package v1

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/authorizer"
	"ibanking-scraper/internal/http/handler/mutasi"
	"ibanking-scraper/internal/http/router"
)

func NewMutasi(uc domain.MutasiUsecase) router.Router {
	r := route.Group("/internet-banking/mutasi")

	r.GET("/", mutasi.FetchMutasiByDate(uc)).Restricted(authorizer.ForAuthenticatedUser())

	return r
}
