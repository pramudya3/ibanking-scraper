package rekening

import (
	"context"
	"fmt"
	"ibanking-scraper/domain"
	"ibanking-scraper/domain/types"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/logger"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const rekeningQuery = `
	SELECT
	  	id,
		tipe_bank,
		rekening,
		pemilik_rekening,
		saldo,
		griyabayar,
		tgl_update
	FROM
		ibanking.saldo_rekening`

type rekeningRepository struct {
	db *sqlx.DB
}

func (c *rekeningRepository) GetByIdAkunBank(ctx context.Context, idAkunBank uint64) ([]*domain.Rekening, error) {
	rekenings := []*domain.Rekening{}

	rows, err := c.db.QueryxContext(ctx, rekeningQuery+` WHERE id_akun_bank=$1`, idAkunBank)
	if err != nil {
		logger.Error(err)
		return nil, errors.Transform(err)
	}

	for rows.Next() {
		tmp := &domain.Rekening{}

		if err = rows.Scan(
			&tmp.ID,
			&tmp.TipeBank,
			&tmp.Rekening,
			&tmp.PemilikRekening,
			&tmp.Saldo,
			&tmp.Griyabayar,
			&tmp.TglUpdate,
		); err != nil {
			logger.Error(err)
			return nil, errors.Transform(err)
		}

		rekenings = append(rekenings, tmp)
	}

	if err = rows.Err(); err != nil {
		logger.Error(err)
		return nil, errors.Transform(err)
	}

	return rekenings, nil
}

func (c *rekeningRepository) Fetch(ctx context.Context) ([]*domain.Rekening, error) {
	data := []*domain.Rekening{}

	rows, err := c.db.QueryxContext(ctx, rekeningQuery)
	if err != nil {
		logger.Error(err)
		return nil, errors.Transform(err)
	}

	for rows.Next() {
		tmp := domain.Rekening{}

		if err = rows.Scan(
			&tmp.ID,
			&tmp.TipeBank,
			&tmp.Rekening,
			&tmp.PemilikRekening,
			&tmp.Saldo,
			&tmp.Griyabayar,
			&tmp.TglUpdate,
		); err != nil {
			logger.Error(err)
			return nil, errors.Transform(err)
		}

		data = append(data, &tmp)
	}

	if err = rows.Err(); err != nil {
		logger.Error(err)
		return nil, errors.Transform(err)
	}

	return data, nil
}

func (c *rekeningRepository) BulkInsertRekOnpay(ctx context.Context, rekenings []*domain.Rekening) error {
	var (
		placeholders []string
		vals         []interface{}
		nullId       types.NullInt
		lastId       int64
		attempt      int
	)

Loop:
	if err := c.db.QueryRowx(`SELECT MAX(id) AS last_id FROM ibanking.saldo_rekening;`).Scan(&nullId); err != nil {
		logger.Error(err)
		return errors.Transform(err)
	}

	if nullId.Valid {
		lastId = nullId.Int64
	}

	for index, rekening := range rekenings {
		lastId = lastId + 1
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			index*7+1,
			index*7+2,
			index*7+3,
			index*7+4,
			index*7+5,
			index*7+6,
			index*7+7,
		))

		vals = append(
			vals,
			lastId,
			rekening.TipeBank,
			rekening.Rekening,
			rekening.PemilikRekening,
			rekening.Saldo,
			false,
			time.Now(),
		)
	}

	tx, err := c.db.Begin()
	if err != nil {
		return errors.Transform(err)
	}

	insertStatement := fmt.Sprintf(`INSERT INTO ibanking.saldo_rekening(
			id,
			tipe_bank,
			rekening,
			pemilik_rekening,
			saldo,
		    griyabayar,
			tgl_update
		) VALUES %s
		ON CONFLICT ON CONSTRAINT saldo_rekening_un
		DO UPDATE SET
			tipe_bank=excluded.tipe_bank,
			rekening=excluded.rekening,
			pemilik_rekening=excluded.pemilik_rekening,
			saldo=excluded.saldo,
		    griyabayar=excluded.griyabayar,
			tgl_update=excluded.tgl_update;`,
		strings.Join(placeholders, ","))
	if _, err = tx.Exec(insertStatement, vals...); err != nil {
		tx.Rollback()
		logger.Error(err)
		if attempt < 5 {
			attempt++
			goto Loop
		} else {
			logger.Error(err)
			return errors.Transform(err)
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return errors.Transform(err)
	}

	return nil
}

func (c *rekeningRepository) BulkInsertRekGriyabayar(ctx context.Context, rekenings []*domain.Rekening) error {
	var (
		placeholders []string
		vals         []interface{}
		nullId       types.NullInt
		lastId       int64
		attempt      int
	)

Loop:
	if err := c.db.QueryRowx(`SELECT MAX(id) AS last_id FROM ibanking.saldo_rekening;`).Scan(&nullId); err != nil {
		logger.Error(err)
		return errors.Transform(err)
	}

	if nullId.Valid {
		lastId = nullId.Int64
	}

	for index, rekening := range rekenings {
		lastId = lastId + 1
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			index*7+1,
			index*7+2,
			index*7+3,
			index*7+4,
			index*7+5,
			index*7+6,
			index*7+7,
		))

		vals = append(
			vals,
			lastId,
			rekening.TipeBank,
			rekening.Rekening,
			rekening.PemilikRekening,
			rekening.Saldo,
			true,
			time.Now(),
		)
	}

	tx, err := c.db.Begin()
	if err != nil {
		return errors.Transform(err)
	}

	insertStatement := fmt.Sprintf(`INSERT INTO ibanking.saldo_rekening(
			id,
			tipe_bank,
			rekening,
			pemilik_rekening,
			saldo,
		    griyabayar,
			tgl_update
		) VALUES %s
		ON CONFLICT ON CONSTRAINT saldo_rekening_un
		DO UPDATE SET
			tipe_bank=excluded.tipe_bank,
			rekening=excluded.rekening,
			pemilik_rekening=excluded.pemilik_rekening,
			saldo=excluded.saldo,
		    griyabayar=excluded.griyabayar,
			tgl_update=excluded.tgl_update;`,
		strings.Join(placeholders, ","))
	if _, err = tx.Exec(insertStatement, vals...); err != nil {
		tx.Rollback()
		logger.Error(err)
		if attempt < 5 {
			attempt++
			goto Loop
		} else {
			logger.Error(err)
			return errors.Transform(err)
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return errors.Transform(err)
	}

	return nil
}

func NewRekeningRepository(db *sqlx.DB) domain.RekeningRepository {
	return &rekeningRepository{
		db: db,
	}
}
