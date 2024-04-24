package mutasi

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

const mutasiQuery = `
	SELECT
		id,
		tgl_entri,
		tgl_bank,
		tipe_bank,
		rekening,
		pemilik_rekening,
		keterangan,
		tipe_mutasi,
		jumlah,
		saldo,
		terklaim,
		griyabayar,
		id_tiket,
		catatan
	FROM
		ibanking.mutasi_bank `

type mutasiRepository struct {
	db *sqlx.DB
}

func (c *mutasiRepository) FetchByRekening(ctx context.Context, rekening string, filter domain.Filter) ([]*domain.Mutasi, error) {
	datas := []*domain.Mutasi{}

	rows, err := c.db.QueryxContext(ctx, `
SELECT mb.id, mb.tgl_entri, mb.tgl_bank, mb.tipe_bank, mb.griyabayar, mb.pemilik_rekening, sr.rekening, mb.keterangan, mb.tipe_mutasi, mb.jumlah, mb.saldo, sr.id_akun_bank 
FROM ibanking.mutasi_bank mb 
LEFT JOIN ibanking.saldo_rekening sr on mb.rekening = sr.rekening
WHERE sr.rekening = $1
ORDER BY mb.tgl_bank ASC `+`LIMIT $2 OFFSET $3`, rekening, filter.Limit(), filter.Offset())
	if err != nil {
		logger.Error(err)
		return nil, errors.Transform(err)
	}

	for rows.Next() {
		tmp := &domain.Mutasi{}

		if err = rows.Scan(
			&tmp.ID,
			&tmp.TglEntri,
			&tmp.TglBank,
			&tmp.TipeBank,
			&tmp.PemilikRekening,
			&tmp.Rekening,
			&tmp.Keterangan,
			&tmp.TipeMutasi,
			&tmp.Jumlah,
			&tmp.Saldo,
			&tmp.Terklaim,
			&tmp.Griyabayar,
			&tmp.IdTiket,
			&tmp.Catatan,
		); err != nil {
			logger.Error(err)
			return nil, errors.Transform(err)
		}

		datas = append(datas, tmp)
	}

	if err = rows.Err(); err != nil {
		logger.Error(err)
		return nil, errors.Transform(err)
	}

	return datas, nil
}

func (c *mutasiRepository) FetchByDate(ctx context.Context, startDate, endDate string, rekening string, isGriyabayar bool, filter domain.Filter) ([]*domain.Mutasi, interface{}, error) {
	data := []*domain.Mutasi{}
	totalRecords := 0

	rows, err := c.db.QueryxContext(ctx, mutasiQuery+fmt.Sprintf(`WHERE 
																				tgl_bank >= '%s'
																				and tgl_bank < '%s'
																				and rekening = '%s'
																				and griyabayar = %t
																				order by tgl_bank desc `, endDate, startDate, rekening, isGriyabayar)+` LIMIT $1 OFFSET $2`, filter.Limit(), filter.Offset())
	if err != nil {
		logger.Error(err)
		return nil, nil, errors.Transform(err)
	}

	for rows.Next() {
		tmp := domain.Mutasi{}

		if err = rows.Scan(
			&tmp.ID,
			&tmp.TglEntri,
			&tmp.TglBank,
			&tmp.TipeBank,
			&tmp.Rekening,
			&tmp.PemilikRekening,
			&tmp.Keterangan,
			&tmp.TipeMutasi,
			&tmp.Jumlah,
			&tmp.Saldo,
			&tmp.Terklaim,
			&tmp.Griyabayar,
			&tmp.IdTiket,
			&tmp.Catatan,
		); err != nil {
			logger.Error(err)
			return nil, nil, errors.Transform(err)
		}

		data = append(data, &tmp)
	}

	if err = rows.Err(); err != nil {
		logger.Error(err)
		return nil, nil, errors.Transform(err)
	}

	if err := c.db.QueryRowContext(ctx, `SELECT COUNT(j) FROM(`+mutasiQuery+`)j`).Scan(&totalRecords); err != nil {
		logger.Error(err)
	}

	meta := filter.BuildMetadata(totalRecords)

	return data, meta, nil
}

func (c *mutasiRepository) Fetch(ctx context.Context, rekening string) ([]*domain.Mutasi, error) {
	data := []*domain.Mutasi{}

	rows, err := c.db.QueryxContext(ctx, mutasiQuery+` WHERE rekening = $1 ORDER BY id DESC`, rekening)
	if err != nil {
		logger.Error(err)
		return nil, errors.Transform(err)
	}

	for rows.Next() {
		tmp := domain.Mutasi{}

		if err = rows.Scan(
			&tmp.ID,
			&tmp.TglEntri,
			&tmp.TglBank,
			&tmp.TipeBank,
			&tmp.PemilikRekening,
			&tmp.Rekening,
			&tmp.Keterangan,
			&tmp.TipeMutasi,
			&tmp.Jumlah,
			&tmp.Saldo,
			&tmp.Terklaim,
			&tmp.Griyabayar,
			&tmp.IdTiket,
			&tmp.Catatan,
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

func (c *mutasiRepository) FetchRekGriyabayar(ctx context.Context) ([]*domain.Mutasi, error) {
	data := []*domain.Mutasi{}

	rows, err := c.db.QueryxContext(ctx, mutasiQuery+` WHERE griyabayar = true`)
	if err != nil {
		logger.Error(err)
		return nil, errors.Transform(err)
	}

	for rows.Next() {
		tmp := domain.Mutasi{}

		if err = rows.Scan(
			&tmp.ID,
			&tmp.TglEntri,
			&tmp.TglBank,
			&tmp.TipeBank,
			&tmp.PemilikRekening,
			&tmp.Rekening,
			&tmp.Keterangan,
			&tmp.TipeMutasi,
			&tmp.Jumlah,
			&tmp.Saldo,
			&tmp.Terklaim,
			&tmp.Griyabayar,
			&tmp.IdTiket,
			&tmp.Catatan,
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

func (c *mutasiRepository) BulkInsertRekOnpay(ctx context.Context, mutasis []*domain.Mutasi) error {
	var (
		placeholders []string
		vals         []interface{}
		nullId       types.NullInt
		lastId       int64
		attempt      int
	)

Loop:
	if err := c.db.QueryRowx(`SELECT MAX(id) AS last_id FROM ibanking.mutasi_bank;`).Scan(&nullId); err != nil {
		logger.Error(err)
		return errors.Transform(err)
	}

	if nullId.Valid {
		lastId = nullId.Int64
	}

	for index, mutasi := range mutasis {
		lastId = lastId + 1
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			index*12+1,
			index*12+2,
			index*12+3,
			index*12+4,
			index*12+5,
			index*12+6,
			index*12+7,
			index*12+8,
			index*12+9,
			index*12+10,
			index*12+11,
			index*12+12,
		))

		vals = append(
			vals,
			lastId,
			time.Now(),
			mutasi.TglBank,
			mutasi.TipeBank,
			mutasi.Rekening,
			mutasi.PemilikRekening,
			mutasi.Keterangan,
			mutasi.TipeMutasi,
			mutasi.Jumlah,
			mutasi.Saldo,
			mutasi.Terklaim,
			false,
		)
	}

	tx, err := c.db.Begin()
	if err != nil {
		return errors.Transform(err)
	}

	insertStatement := fmt.Sprintf(`INSERT INTO ibanking.mutasi_bank(
                id,
				tgl_entri,
				tgl_bank,
				tipe_bank,
				rekening,
				pemilik_rekening,
				keterangan,
				tipe_mutasi,
				jumlah,
				saldo,
                terklaim,
                griyabayar
		) VALUES %s
		ON CONFLICT ON CONSTRAINT mutasi_bank_un
		DO UPDATE SET
			tgl_entri=excluded.tgl_entri,
		    tgl_bank=excluded.tgl_bank,
		    keterangan=excluded.keterangan,
		    jumlah=excluded.jumlah,
		    saldo=excluded.saldo;`,
		strings.Join(placeholders, ","))
	_, err = tx.Exec(insertStatement, vals...)
	if err != nil {
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

func (c *mutasiRepository) BulkInsertRekGriyabayar(ctx context.Context, mutasis []*domain.Mutasi) error {
	var (
		placeholders []string
		vals         []interface{}
		nullId       types.NullInt
		lastId       int64
		attempt      int
	)

Loop:
	if err := c.db.QueryRowx(`SELECT MAX(id) AS last_id FROM ibanking.mutasi_bank;`).Scan(&nullId); err != nil {
		logger.Error(err)
		return errors.Transform(err)
	}

	if nullId.Valid {
		lastId = nullId.Int64
	}

	for index, mutasi := range mutasis {
		lastId = lastId + 1
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			index*12+1,
			index*12+2,
			index*12+3,
			index*12+4,
			index*12+5,
			index*12+6,
			index*12+7,
			index*12+8,
			index*12+9,
			index*12+10,
			index*12+11,
			index*12+12,
		))

		vals = append(
			vals,
			lastId,
			time.Now(),
			mutasi.TglBank,
			mutasi.TipeBank,
			mutasi.Rekening,
			mutasi.PemilikRekening,
			mutasi.Keterangan,
			mutasi.TipeMutasi,
			mutasi.Jumlah,
			mutasi.Saldo,
			mutasi.Terklaim,
			false,
		)
	}

	tx, err := c.db.Begin()
	if err != nil {
		return errors.Transform(err)
	}

	insertStatement := fmt.Sprintf(`INSERT INTO ibanking.mutasi_bank(
                id,
				tgl_entri,
				tgl_bank,
				tipe_bank,
				rekening,
				pemilik_rekening,
				keterangan,
				tipe_mutasi,
				jumlah,
				saldo,
                terklaim,
                griyabayar
		) VALUES %s 
		ON CONFLICT ON CONSTRAINT mutasi_bank_un
		DO UPDATE SET
			tgl_entri=excluded.tgl_entri,
		    tgl_bank=excluded.tgl_bank,
		    keterangan=excluded.keterangan,
		    jumlah=excluded.jumlah,
		    saldo=excluded.saldo;`,
		strings.Join(placeholders, ","))

	_, err = tx.Exec(insertStatement, vals...)
	if err != nil {
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

func NewMutasiRepository(db *sqlx.DB) domain.MutasiRepository {
	return &mutasiRepository{
		db: db,
	}
}
