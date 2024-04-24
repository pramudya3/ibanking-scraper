package types

import (
	"database/sql/driver"
	"encoding/json"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/pkg/constant"
	"time"
)

type PGTime struct {
	time.Time
}

func (tm *PGTime) Set(value string) error {
	var err error
	tm.Time, err = time.Parse(constant.LayoutTime, value)
	if err != nil {
		return err
	}
	return nil
}

func (tm PGTime) String() string {
	return tm.Format(constant.LayoutTime)
}

// Scan implements the Scanner interface for PGTime
func (tm *PGTime) Scan(value interface{}) error {
	if v, ok := value.(string); ok {
		return tm.Set(v)
	}

	return errors.Unknown.New("PGTime: unknown time error")
}

// Value implements the driver Valuer interface.
func (tm PGTime) Value() (driver.Value, error) {
	return tm.String(), nil
}

// MarshalJSON for PGTime
func (tm PGTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(tm.String())
}

// UnmarshalJSON for PGTime
func (tm *PGTime) UnmarshalJSON(data []byte) error {
	var t string
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	if err := tm.Set(t); err != nil {
		return err
	}

	return nil
}
