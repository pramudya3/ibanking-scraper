package scraper

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/saved_browser"
	"ibanking-scraper/internal/scraper"
	"ibanking-scraper/utils"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/makelifemorefun/cron"
)

func StartScrepeAll(
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

		for _, bankAccount := range bankAccounts {
			passwordDecoded := utils.DecodePassword(bankAccount.Password.String)
			bankAccount.Password.String = passwordDecoded
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
				go ibanking.StartScrape()
			}
		}

		return c.NoContent(http.StatusOK)
	}
}
