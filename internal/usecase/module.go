package usecase

import (
	"ibanking-scraper/internal/usecase/bank_account"
	"ibanking-scraper/internal/usecase/iam"
	"ibanking-scraper/internal/usecase/log"
	"ibanking-scraper/internal/usecase/mutasi"
	"ibanking-scraper/internal/usecase/rekening"
	"ibanking-scraper/internal/usecase/user"

	"go.uber.org/fx"
)

var Module = fx.Provide(
	mutasi.NewMutasiUsecase,
	rekening.NewRekeningUsecase,
	bank_account.NewBankAccountUsecase,
	log.NewLogUsecase,
	user.NewUserUsecase,
	iam.NewIAMUsecase,
)
