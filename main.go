package main

import (
	"ibanking-scraper/cmd"
	"ibanking-scraper/internal/logger"
	"os"

	_ "github.com/jackc/pgx/stdlib"
)

/*func main() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s TimeZone=Asia/Jakarta",
		"mpndb.cr6wwc6xwxel.ap-southeast-3.rds.amazonaws.com",
		5432,
		"postgres",
		"mpn_internetbanking",
		"Asdfgvcxz",
		"disable",
	)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		logger.Debug("could not connect to database:", err)
	}

	db.SetMaxOpenConns(50)
 	db.SetMaxIdleConns(10)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(10*time.Second))
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		logger.Debug("could not ping to database:", err)
	}

	rekeningRepo := repository.NewRekeningRepository(db)
	rekeningUseCase := usecase.NewRekeningUsecase(rekeningRepo)
	mutasiRepo := repository.NewMutasiRepository(db)
	mutasiUseCase := usecase.NewMutasiUsecase(mutasiRepo)

	err = bripersonal.FetchBRIPersonal(rekeningUseCase, mutasiUseCase)
	if err != nil {
		logger.Debug("failed to scrape:", err)
	}
	bripersonal.FetchBRIPersonal()
}*/

func main() {
	if err := os.Setenv("TZ", "Asia/Jakarta"); err != nil {
		logger.Error("err while set timezone : ", err)
	}
	cmd.Execute()
}
