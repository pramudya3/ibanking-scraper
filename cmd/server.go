package cmd

import (
	"context"
	"fmt"
	"ibanking-scraper/cmd/dependency"
	"ibanking-scraper/config"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/logger"
	"ibanking-scraper/internal/saved_browser"
	"ibanking-scraper/internal/scraper"
	"ibanking-scraper/utils"
	"time"

	"github.com/makelifemorefun/cron"

	"github.com/indrasaputra/hashids"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewCmdServer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apiserver",
		Short: "Run mpn API server",
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		opts := fx.Options(
			dependency.Init(),
			dependency.Inject(),
			fx.Invoke(registerServerHooks),
		)

		app := fx.New(opts)
		app.Run()
	}
	return cmd
}

type ServerParams struct {
	fx.In

	AppConfig *config.Config
	AppLogger *logger.Logger

	EchoServer    *echo.Echo
	Cron          *cron.Cron
	UcLogs        domain.LogUsecase
	UcRekening    domain.RekeningUsecase
	UcMutasi      domain.MutasiUsecase
	UcBankAccount domain.BankAccountUsecase
}

func registerServerHooks(lc fx.Lifecycle, params ServerParams) {
	log := params.AppLogger.Logger()

	hash, err := hashids.NewHashID(uint(params.AppConfig.Hashid.MinLength), params.AppConfig.Hashid.Salt)
	if err != nil {
		log.Sugar().Error(err)
	}
	hashids.SetHasher(hash)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func(log *zap.Logger) {
				for {
					time.Sleep(600 * time.Second)
				}
			}(log)
			go startHTTPServer(log, params.EchoServer, params.AppConfig.ServerAddr, params.AppConfig.ServerPort)
			go autoStart(params.Cron, params.UcLogs, params.UcRekening, params.UcMutasi, params.UcBankAccount)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return stopHTTPServer(ctx, log, params.EchoServer)
		},
	})
}

func startHTTPServer(log *zap.Logger, server *echo.Echo, address string, port int) {
	var routes []string
	for _, route := range server.Routes() {
		routeName := fmt.Sprintf("%s %s", route.Method, route.Path)
		log.With(zap.String("name", route.Name), zap.String("route", routeName)).Info("HTTP Route registration")
		routes = append(routes, routeName)
	}
	log.With(zap.Int("available_routes", len(routes))).
		Info(fmt.Sprintf("Starting HTTP server on port %s:%d", address, port))
	if err := server.Start(fmt.Sprintf("%s:%d", address, port)); err != nil {
		log.With(zap.Error(err)).Error("failed to start HTTP server")
	}
}

func stopHTTPServer(ctx context.Context, log *zap.Logger, server *echo.Echo) error {
	log.Info("Shutting down the HTTP server")
	if err := server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

func autoStart(
	cron *cron.Cron,
	ucLogs domain.LogUsecase,
	ucRekening domain.RekeningUsecase,
	ucMutasi domain.MutasiUsecase,
	ucBankAccount domain.BankAccountUsecase,
) {
	bankAccounts, err := ucBankAccount.Fetch(context.Background())
	if err != nil {
		logger.Error(err)
		return
	}

	for _, bankAccount := range bankAccounts {
		passwordDecoded := utils.DecodePassword(bankAccount.Password.String)
		bankAccount.Password.String = passwordDecoded
		if bankAccount.Aktif {
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

			go ibanking.StartScrape()
		}
	}
}
