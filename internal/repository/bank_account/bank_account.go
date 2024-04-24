package bank_account

import (
	"context"
	"database/sql"
	"errors"
	"ibanking-scraper/domain"
	"ibanking-scraper/domain/types"
	"ibanking-scraper/internal/logger"
	"ibanking-scraper/pkg/constant"
	"time"

	"github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

const bankAccountQuery = `
	SELECT
		id,
		tipe_akun,
		company_id,
		user_id,
		password,
		rek_onpay,
		rek_griyabayar,
		interval_cek,
		total_cek_hari,
		jam_aktif_start,
		jam_aktif_end,	
		auto_logout,
		aktif,
		token,
		status_login,
		tgl_aktivitas,
		tgl_update
	FROM
		ibanking.akun_bank `

type bankAccountRepository struct {
	db *sqlx.DB
}

func (c *bankAccountRepository) FetchToken(ctx context.Context, id uint64) (pq.StringArray, error) {
	tokens := pq.StringArray{}
	if err := c.db.QueryRowContext(ctx, `select token from ibanking.akun_bank where id = $1`, id).Scan(&tokens); err != nil {
		logger.Error(err)
	}
	return tokens, nil
}

func (c *bankAccountRepository) GetToken(ctx context.Context, id uint64) (types.NullString, error) {
	var token types.NullString

	if err := c.db.QueryRowContext(ctx, `select token[1] from ibanking.akun_bank where id = $1;`, id).Scan(&token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Error(sql.ErrNoRows)
			return token, err
		}
		logger.Error(err)
		return token, err
	}
	return token, nil
}

func (c *bankAccountRepository) DeleteToken(ctx context.Context, token string, id uint64) error {
	if _, err := c.db.ExecContext(ctx, `UPDATE ibanking.akun_bank SET token = array_remove(token, $1) where id=$2;`, token, id); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (c *bankAccountRepository) AddToken(ctx context.Context, id uint64, token pq.StringArray) error {
	if _, err := c.db.ExecContext(ctx,
		`UPDATE ibanking.akun_bank
				SET token = (SELECT array_agg(DISTINCT unnest) FROM unnest($1::text[]) AS unnest)
				WHERE id = $2;`,
		token,
		id,
	); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (c *bankAccountRepository) Fetch(ctx context.Context) ([]*domain.BankAccount, error) {
	var data []*domain.BankAccount

	rows, err := c.db.QueryxContext(ctx, bankAccountQuery+`ORDER BY tipe_akun ASC`)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	for rows.Next() {
		tmp := &domain.BankAccount{}

		var jamAktifStartString string
		var jamAktifEndString string

		if err = rows.Scan(
			&tmp.ID,
			&tmp.TipeAkun,
			&tmp.CompanyId,
			&tmp.UserId,
			&tmp.Password,
			&tmp.RekOnpay,
			&tmp.RekGriyabayar,
			&tmp.IntervalCek,
			&tmp.TotalCekHari,
			&jamAktifStartString,
			&jamAktifEndString,
			&tmp.AutoLogout,
			&tmp.Aktif,
			&tmp.Token,
			&tmp.StatusLogin,
			&tmp.TglAktivitas,
			&tmp.TglUpdate,
		); err != nil {
			logger.Error(err)
			return nil, err
		}

		jamAktifStartTime, _ := time.Parse(constant.LayoutTimeWithFloat, jamAktifStartString)
		jamAktifEndTime, _ := time.Parse(constant.LayoutTimeWithFloat, jamAktifEndString)
		tmp.JamAktifStart.Time = jamAktifStartTime
		tmp.JamAktifEnd.Time = jamAktifEndTime

		data = append(data, tmp)
	}

	if err = rows.Err(); err != nil {
		logger.Error(err)
		return nil, err
	}
	return data, nil
}

func (c *bankAccountRepository) FindByID(ctx context.Context, id uint64) (*domain.BankAccount, error) {
	tmp := &domain.BankAccount{}

	if err := c.db.QueryRowxContext(ctx, bankAccountQuery+` WHERE id=$1`, id).Scan(
		&tmp.ID,
		&tmp.TipeAkun,
		&tmp.CompanyId,
		&tmp.UserId,
		&tmp.Password,
		&tmp.RekOnpay,
		&tmp.RekGriyabayar,
		&tmp.IntervalCek,
		&tmp.TotalCekHari,
		&tmp.JamAktifStart,
		&tmp.JamAktifEnd,
		&tmp.AutoLogout,
		&tmp.Aktif,
		&tmp.Token,
		&tmp.StatusLogin,
		&tmp.TglAktivitas,
		&tmp.TglUpdate,
	); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Error(err)
			return nil, err
		} else {
			return nil, sql.ErrNoRows
		}
	}

	return tmp, nil
}

func (c *bankAccountRepository) Create(ctx context.Context, bankAccount *domain.BankAccount) error {
	if err := c.db.QueryRowContext(ctx,
		`INSERT INTO ibanking.akun_bank(
			tipe_akun,
			company_id,
			user_id,
			password,
            rek_onpay,
			rek_griyabayar,
			interval_cek,
			total_cek_hari,
			auto_logout,
			jam_aktif_start,
			jam_aktif_end,
            aktif,
			tgl_update
		)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`,
		bankAccount.TipeAkun,
		bankAccount.CompanyId,
		bankAccount.UserId,
		bankAccount.Password,
		bankAccount.RekOnpay,
		bankAccount.RekGriyabayar,
		bankAccount.IntervalCek,
		bankAccount.TotalCekHari,
		bankAccount.AutoLogout,
		bankAccount.JamAktifStart,
		bankAccount.JamAktifEnd,
		bankAccount.Aktif,
		time.Now(),
	).Scan(
		&bankAccount.ID,
	); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (c *bankAccountRepository) Update(ctx context.Context, bankAccount *domain.BankAccount, id uint64) error {
	if _, err := c.db.ExecContext(ctx,
		`UPDATE
			ibanking.akun_bank
		SET
			tipe_akun=$1,
			company_id=$2,
			user_id=$3,
			password=$4,
			rek_onpay=$5,
			rek_griyabayar=$6,
			interval_cek=$7,
			total_cek_hari=$8,
			auto_logout=$9,
			jam_aktif_start=$10,
			jam_aktif_end=$11,
			aktif=$12,
			tgl_update=$13
		WHERE id=$14;`,
		bankAccount.TipeAkun,
		bankAccount.CompanyId,
		bankAccount.UserId,
		bankAccount.Password.String,
		bankAccount.RekOnpay,
		bankAccount.RekGriyabayar,
		bankAccount.IntervalCek,
		bankAccount.TotalCekHari,
		bankAccount.AutoLogout,
		bankAccount.JamAktifStart,
		bankAccount.JamAktifEnd,
		bankAccount.Aktif,
		time.Now(),
		id,
	); err != nil {
		if err != sql.ErrNoRows {
			logger.Error(err)
			return err
		} else {
			return sql.ErrNoRows
		}
	}

	return nil
}

func (c *bankAccountRepository) Delete(ctx context.Context, id uint64) error {
	if _, err := c.db.ExecContext(ctx, `DELETE FROM ibanking.akun_bank WHERE id=$1;`, id); err != nil {
		if err != sql.ErrNoRows {
			logger.Error(err)
			return err
		} else {
			return sql.ErrNoRows
		}
	}
	return nil
}

func (c *bankAccountRepository) UpdateLoginStatus(ctx context.Context, bankAccount *domain.BankAccount, id uint64) error {
	if _, err := c.db.ExecContext(ctx,
		`UPDATE
			ibanking.akun_bank
		SET
			status_login=$1,
			tgl_aktivitas=$2
		WHERE id=$3;`,
		bankAccount.StatusLogin,
		time.Now(),
		id,
	); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func NewBankAccountRepository(db *sqlx.DB) domain.BankAccountRepository {
	return &bankAccountRepository{
		db: db,
	}
}
