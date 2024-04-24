package log

import (
	"context"
	"ibanking-scraper/domain"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/logger"
	"time"

	"github.com/jmoiron/sqlx"
)

const logQuery = `
	SELECT
		id,
		tgl_entri, 
		akun_bank_id, 
		tipe,
		keterangan
	FROM
		public.ibanking_logs `

type logRepository struct {
	db *sqlx.DB
}

func (c *logRepository) FetchByAkunBank(ctx context.Context, akunBank uint64, filter domain.Filter) ([]*domain.Log, interface{}, error) {
	data := []*domain.Log{}
	totalRecords := 0

	rows, err := c.db.QueryContext(ctx, logQuery+`WHERE akun_bank_id=$1 ORDER BY tgl_entri DESC LIMIT $2 OFFSET $3`, akunBank, filter.Limit(), filter.Offset())
	if err != nil {
		logger.Error(err)
		return nil, nil, errors.Transform(err)
	}
	for rows.Next() {
		tmp := domain.Log{}

		if err = rows.Scan(
			&tmp.ID,
			&tmp.TglEntri,
			&tmp.AkunBankId,
			&tmp.Tipe,
			&tmp.Keterangan,
		); err != nil {
			logger.Error(err)
			return nil, nil, errors.Transform(err)
		}

		data = append(data, &tmp)
	}

	if err := rows.Err(); err != nil {
		logger.Error(err)
		return nil, nil, errors.Transform(err)
	}

	if err := c.db.QueryRowContext(ctx, `SELECT COUNT(j) FROM(`+logQuery+` WHERE akun_bank_id=$1)j`, akunBank).Scan(&totalRecords); err != nil {
		logger.Error(err)
		return nil, nil, errors.Transform(err)
	}

	meta := filter.BuildMetadata(totalRecords)

	return data, meta, nil
}

func (c *logRepository) Create(ctx context.Context, dataLog *domain.Log) error {
	if err := c.db.QueryRowContext(ctx,
		` INSERT INTO public.ibanking_logs(
			tgl_entri, 
			akun_bank_id,
        	tipe,
			keterangan
		)
		VALUES($1, $2, $3, $4)
		RETURNING id`,
		time.Now(),
		dataLog.AkunBankId,
		dataLog.Tipe,
		dataLog.Keterangan,
	).Scan(&dataLog.ID); err != nil {
		logger.Error(err)
		return errors.Transform(err)
	}
	return nil
}

func NewLogRepository(db *sqlx.DB) domain.LogRepository {
	return &logRepository{
		db: db,
	}
}
