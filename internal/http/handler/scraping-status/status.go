package scraping_status

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/http/response"
	"ibanking-scraper/internal/saved_browser"
	"net/http"

	"github.com/indrasaputra/hashids"
	"github.com/labstack/echo/v4"
)

type StatusResponse struct {
	ID          hashids.ID         `json:"id"`
	TipeAkun    domain.AccountType `json:"tipe"`
	StatusLogin domain.LoginStatus `json:"statusLogin"`
}

func GetStatusBankAccount(uc domain.BankAccountUsecase) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		ctx := c.Request().Context()

		bankAccounts, err := uc.Fetch(ctx)
		if err != nil {
			return err
		}

		browser := saved_browser.GetSavedBrowser()
		var resps []*StatusResponse
		for _, bankAccount := range bankAccounts {
			resp := &StatusResponse{
				ID:       bankAccount.ID,
				TipeAkun: bankAccount.TipeAkun,
			}
			page := browser.Page[uint64(bankAccount.ID)]
			if page != nil {
				resp.StatusLogin = bankAccount.StatusLogin
			} else {
				resp.StatusLogin = domain.BelumLogin
			}
			resps = append(resps, resp)
		}

		return c.JSON(http.StatusOK, response.NewSuccess(resps, response.EmptyMeta{}))
	}
}
