package v1

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/authorizer"
	scraping_status "ibanking-scraper/internal/http/handler/scraping-status"
	"ibanking-scraper/internal/http/router"
)

func NewBankStatusRoute(c domain.BankAccountUsecase) router.Router {

	r := route.Group("/internet-banking/status")

	r.GET("/", scraping_status.GetStatusBankAccount(c)).Restricted(authorizer.ForAuthenticatedUser())

	return r
}
