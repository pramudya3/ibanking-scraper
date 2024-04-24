package v1

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/authorizer"
	"ibanking-scraper/internal/http/handler/bank_account"
	"ibanking-scraper/internal/http/middleware"
	"ibanking-scraper/internal/http/router"

	"github.com/labstack/echo/v4"
)

func NewBankAccountRoute(c domain.BankAccountUsecase) router.Router {
	middJSON := middleware.WithContentType(echo.MIMEApplicationJSON)

	r := route.Group("/internet-banking")

	r.GET("/:id/", bank_account.GetBankAccount(c)).Restricted(authorizer.ForAuthenticatedUser())
	r.POST("/", bank_account.AddBankAccount(c), middJSON).Restricted(authorizer.ForAuthenticatedUser())
	r.PUT("/:id/", bank_account.UpdateBankAccount(c), middJSON).Restricted(authorizer.ForAuthenticatedUser())
	r.DELETE("/:id/", bank_account.DeleteBankAccount(c)).Restricted(authorizer.ForAuthenticatedUser())

	return r
}
