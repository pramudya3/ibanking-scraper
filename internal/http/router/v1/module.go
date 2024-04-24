package v1

import (
	"ibanking-scraper/internal/http/router"

	"go.uber.org/fx"
)

var route = router.New("/api/v1")
var Module = fx.Options(
	router.RegisterFx(NewBankAccountRoute),
	router.RegisterFx(NewLogRoute),
	router.RegisterFx(NewScraperRoute),
	router.RegisterFx(NewHealthCheckRoute),
	router.RegisterFx(NewBankStatusRoute),
	router.RegisterFx(NewInsertTokenRoute),
	router.RegisterFx(NewMutasi),
)
