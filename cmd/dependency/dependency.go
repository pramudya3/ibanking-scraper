package dependency

import (
	"ibanking-scraper/config"
	cron_initiate "ibanking-scraper/internal/cron"
	"ibanking-scraper/internal/db"
	"ibanking-scraper/internal/decoder"
	"ibanking-scraper/internal/http"
	"ibanking-scraper/internal/logger"
	"ibanking-scraper/internal/repository"
	"ibanking-scraper/internal/types"
	"ibanking-scraper/internal/usecase"
	"time"

	"go.uber.org/fx"
)

type Out struct {
	fx.Out

	Config  *config.Config
	Logger  *logger.Logger
	Timeout types.ServerContextTimeoutDuration
}

func Init() fx.Option {
	var out Out

	conf := config.Load()
	logger := logger.Setup()

	out.Config = conf
	out.Logger = logger
	out.Timeout = types.ServerContextTimeoutDuration(time.Duration(conf.ServerTimeoutContext) * time.Second)
	return fx.Options(
		fx.Logger(logger),
		fx.Provide(func() Out {
			return out
		}),
	)
}

func Inject() fx.Option {
	return fx.Options(
		db.Module,
		repository.Module,
		usecase.Module,
		http.Module,
		cron_initiate.Module,
		decoder.Module,
	)
}
