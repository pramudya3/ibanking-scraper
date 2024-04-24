package scraper

import (
	"context"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/logger"
	"ibanking-scraper/internal/saved_browser"
	"ibanking-scraper/internal/scraper"
	"ibanking-scraper/utils"
	"net/http"

	"github.com/indrasaputra/hashids"
	"github.com/labstack/echo/v4"
	"github.com/makelifemorefun/cron"
)

func GetStartScraper(
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

		savedBrowser := saved_browser.GetSavedBrowser()

		bankAccount, err := ucBankAccount.FindByID(context.Background(), uint64(id))
		if err != nil {
			logger.Error(err)
			return err
		}
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
			go ibanking.StartScrape()
		} else {
			logger.Debug("Akun tidak aktif")
		}

		return c.NoContent(http.StatusOK)
	}
}
