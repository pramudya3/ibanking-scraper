package errors

import (
	"database/sql"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

// GetType returns the error type
func GetType(err error) ErrorType {
	if appErr, ok := err.(applicationError); ok {
		return appErr.errorType
	}

	return NoType
}

// Transform is transforming low-level error to application-level error
func Transform(err error) error {
	// if nil, stay nil
	if err == nil {
		return nil
	}

	// if already an application-level error, stay as-is
	if appErr, ok := err.(applicationError); ok {
		return appErr
	}

	// if postgresql-related error, abstract it with application-level error
	if err := wrapPgErr(err); err != nil {
		return err
	}

	// if sql-related error
	if err := sqlErr(err); err != nil {
		return err
	}

	// if unknown error, wrap the error with the unknown application-level error
	return Unknown.Wrap(err, "Unknown error")
}

func wrapPgErr(err error) error {
	if pgerr, ok := err.(pgx.PgError); ok {
		switch pgerr.Code {
		case pgerrcode.UniqueViolation:
			return Conflict.Wrap(err, "requested resource already exist")
		default:
			return InternalServerError.Wrap(err, "[pg] Something goes wrong when interacting with database")
		}
	}

	return nil
}

func sqlErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrResourceNotFound
	}

	return nil
}
