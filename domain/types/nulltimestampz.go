package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"time"
)

// NullBool is an alias for sql.NullBool data type
type NullTimestampz sql.NullTime

// Scan implements the Scanner interface for NullBool
func (nt *NullTimestampz) Scan(value interface{}) error {
	var b sql.NullTime
	if err := b.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		nt.Valid = false
	} else {
		nt.Time = b.Time
		nt.Valid = true
	}

	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTimestampz) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// MarshalJSON for NullBool
func (nt NullTimestampz) MarshalJSON() ([]byte, error) {
	if nt.Valid {
		return json.Marshal(nt.Time.Unix())
	}
	return json.Marshal(nil)
}

// UnmarshalJSON for NullBool
func (nt *NullTimestampz) UnmarshalJSON(data []byte) error {
	var b interface{}
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}

	switch v := b.(type) {
	case int64:
		nt.Valid = true
		nt.Time = time.Unix(v, 0)
	case time.Time:
		nt.Valid = true
		nt.Time = v
	}

	return nil
}
