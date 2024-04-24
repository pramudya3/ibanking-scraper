package types

import (
	"database/sql/driver"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/internal/logger"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	plaintext *string
	hash      string
}

func (p *Password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.MinCost)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = string(hash)

	return nil
}

func (p *Password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			logger.Debug("not match:", err)
			return false, err
		}
	}

	return true, nil
}

// Scan implements the Scanner interface for Password
func (p *Password) Scan(value interface{}) error {
	p.hash = value.(string)
	return nil
}

// Value implements the driver Valuer Password.
func (p Password) Value() (driver.Value, error) {
	return p.hash, nil
}
