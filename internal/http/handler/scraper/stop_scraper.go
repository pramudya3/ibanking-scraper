package scraper

import (
	"context"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/saved_browser"
	"ibanking-scraper/internal/scraper"
	"net/http"

	"github.com/indrasaputra/hashids"
	"github.com/labstack/echo/v4"
	"github.com/makelifemorefun/cron"
)

func GetStopScraper(
	cron *cron.Cron,
	ucLogs domain.LogUsecase,
	ucRekening domain.RekeningUsecase,
	ucMutasi domain.MutasiUsecase,
	ucBankAccount domain.BankAccountUsecase,
) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		str := c.Param("id")
		id, err := hashids.DecodeHash([]byte(str))
		if err != nil {
			return errors.ErrResourceNotFound
		}

		bankAccount, err := ucBankAccount.FindByID(context.Background(), uint64(id))
		if err != nil {
			return err
		}

		savedBrowser := saved_browser.GetSavedBrowser()
		ibanking := scraper.Ibanking{
			C:             cron,
			SavedBrowser:  savedBrowser,
			BankAccount:   bankAccount,
			UcLogs:        ucLogs,
			UcRekening:    ucRekening,
			UcMutasi:      ucMutasi,
			UcBankAccount: ucBankAccount,
		}
		go ibanking.StopScrape()

		return c.NoContent(http.StatusOK)
	}
}
