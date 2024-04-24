package repository

import (
	"ibanking-scraper/internal/repository/bank_account"
	"ibanking-scraper/internal/repository/iam"
	"ibanking-scraper/internal/repository/log"
	"ibanking-scraper/internal/repository/mutasi"
	"ibanking-scraper/internal/repository/rekening"
	"ibanking-scraper/internal/repository/user"

	"go.uber.org/fx"
)

var Module = fx.Provide(
	mutasi.NewMutasiRepository,
	rekening.NewRekeningRepository,
	bank_account.NewBankAccountRepository,
	log.NewLogRepository,
	user.NewUserRepository,
	iam.NewIAMRepository,
)
