package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"reflect"
)

// NullFloat is an alias for sql.NullBool data type
type NullFloat sql.NullFloat64

// Scan implements the Scanner interface for NullFloat
func (ni *NullFloat) Scan(value interface{}) error {
	var i sql.NullFloat64
	if err := i.Scan(value); err != nil {
		return err
	}

	// if nil then make Valid false
	if reflect.TypeOf(value) == nil {
		*ni = NullFloat{i.Float64, false}
	} else {
		*ni = NullFloat{i.Float64, true}
	}

	return nil
}

// Value implements the driver Valuer interface.
func (ni NullFloat) Value() (driver.Value, error) {
	if !ni.Valid {
		return nil, nil
	}
	return ni.Float64, nil
}

// MarshalJSON for NullFloat
func (ni NullFloat) MarshalJSON() ([]byte, error) {
	if ni.Valid {
		return json.Marshal(ni.Float64)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON for NullFloat
func (ni *NullFloat) UnmarshalJSON(data []byte) error {
	var i *float64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}

	if i != nil {
		ni.Valid = true
		ni.Float64 = *i
	} else {
		ni.Valid = false
	}
	return nil
}
