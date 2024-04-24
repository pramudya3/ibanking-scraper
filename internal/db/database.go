package db

import (
	"context"
	"fmt"
	"ibanking-scraper/config"
	"ibanking-scraper/internal/logger"
	"ibanking-scraper/internal/types"
	"time"

	_ "github.com/jackc/pgx/stdlib" // pgx driver
	"github.com/jmoiron/sqlx"
)

func NewDatabase(conf *config.Config, timeout types.ServerContextTimeoutDuration) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s TimeZone=Asia/Jakarta",
		conf.Database.DatabaseHost,
		conf.Database.DatabasePort,
		conf.Database.DatabaseUser,
		conf.Database.DatabaseName,
		conf.Database.DatabasePassword,
		conf.Database.DatabaseSSL,
	)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	db.SetMaxOpenConns(conf.Database.DatabaseMaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(conf.Database.DatabaseConnMaxlifetime) * time.Second)
	db.SetMaxIdleConns(conf.Database.DatabaseMaxIdleConns)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(timeout))
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		logger.Error(err)
		return nil, err
	}

	return db, nil
}
