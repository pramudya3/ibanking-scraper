package v1

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/authorizer"
	"ibanking-scraper/internal/http/handler/scraper"
	"ibanking-scraper/internal/http/router"

	"github.com/makelifemorefun/cron"
)

func NewScraperRoute(
	cron *cron.Cron,
	ucLogs domain.LogUsecase,
	ucRekening domain.RekeningUsecase,
	ucMutasi domain.MutasiUsecase,
	ucBankAccount domain.BankAccountUsecase,
) router.Router {
	r := route.Group("/internet-banking")

	r.GET("/start-scrape/:id/", scraper.GetStartScraper(cron, ucLogs, ucRekening, ucMutasi, ucBankAccount)).Restricted(authorizer.ForAuthenticatedUser())
	r.GET("/stop-scrape/:id/", scraper.GetStopScraper(cron, ucLogs, ucRekening, ucMutasi, ucBankAccount)).Restricted(authorizer.ForAuthenticatedUser())
	r.GET("/start-scrape-all/", scraper.StartScrepeAll(cron, ucLogs, ucRekening, ucMutasi, ucBankAccount)).Restricted(authorizer.ForAuthenticatedUser())
	r.GET("/stop-scrape-all/", scraper.StopScrepeAll(cron, ucLogs, ucRekening, ucMutasi, ucBankAccount)).Restricted(authorizer.ForAuthenticatedUser())

	return r
}
