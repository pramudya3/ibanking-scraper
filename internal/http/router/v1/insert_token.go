package v1

import (
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/authorizer"
	"ibanking-scraper/internal/http/handler/bank_account"
	"ibanking-scraper/internal/http/middleware"
	"ibanking-scraper/internal/http/router"

	"github.com/labstack/echo/v4"
)

func NewInsertTokenRoute(c domain.BankAccountUsecase) router.Router {
	middJSON := middleware.WithContentType(echo.MIMEApplicationJSON)

	r := route.Group("/internet-banking/insert-token")

	r.POST("/:id/", bank_account.InsertToken(c), middJSON).Restricted(authorizer.ForAuthenticatedUser())
	return r
}
