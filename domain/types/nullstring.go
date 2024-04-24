package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"reflect"
)

// NullBool is an alias for sql.NullBool data type
type NullString sql.NullString

// Scan implements the Scanner interface for NullBool
func (ns *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*ns = NullString{s.String, false}
	} else {
		*ns = NullString{s.String, true}
	}

	return nil
}

// Value implements the driver Valuer interface.
func (nt NullString) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.String, nil
}

// MarshalJSON for NullString
func (nt NullString) MarshalJSON() ([]byte, error) {
	if nt.Valid {
		return json.Marshal(nt.String)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON for NullString
func (nt *NullString) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s != nil {
		nt.Valid = true
		nt.String = *s
	} else {
		nt.Valid = false
	}
	return nil
}
