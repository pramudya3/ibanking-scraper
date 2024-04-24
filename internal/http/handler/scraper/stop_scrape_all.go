package scraper

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/saved_browser"
	"ibanking-scraper/internal/scraper"
	"ibanking-scraper/utils"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/makelifemorefun/cron"
)

func StopScrepeAll(
	cron *cron.Cron,
	ucLogs domain.LogUsecase,
	ucRekening domain.RekeningUsecase,
	ucMutasi domain.MutasiUsecase,
	ucBankAccount domain.BankAccountUsecase,
) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		ctx := c.Request().Context()
		savedBrowser := saved_browser.GetSavedBrowser()

		bankAccounts, err := ucBankAccount.Fetch(ctx)
		if err != nil {
			errors.InternalServerError.New(err.Error())
		}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, bankAccount := range bankAccounts {
				if bankAccount.Aktif {
					ibanking := scraper.Ibanking{
						C:             cron,
						SavedBrowser:  savedBrowser,
						UcLogs:        ucLogs,
						UcRekening:    ucRekening,
						UcMutasi:      ucMutasi,
						UcBankAccount: ucBankAccount,
						BankAccount:   bankAccount,
					}
					passwordDecoded := utils.DecodePassword(bankAccount.Password.String)
					bankAccount.Password.String = passwordDecoded
					go ibanking.StopScrape()
				}
			}
		}()
		wg.Wait()
		return c.NoContent(http.StatusOK)
	}
}
