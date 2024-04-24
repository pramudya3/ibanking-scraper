package types

import (
	"database/sql/driver"
	"encoding/json"
	"ibanking-scraper/internal/errors"
	"ibanking-scraper/pkg/constant"
	"time"
)

type PGDate struct {
	time.Time
}

func (d *PGDate) Set(value string) error {
	var err error
	d.Time, err = time.Parse(constant.LayoutDateISO8601, value)
	if err != nil {
		return err
	}
	return nil
}

func (d PGDate) String() string {
	return d.Format(constant.LayoutDateISO8601)
}

// Scan implements the Scanner interface for PGDate
func (d *PGDate) Scan(value interface{}) error {
	if v, ok := value.(time.Time); ok {
		d.Time = v
		return nil
	}

	return errors.Unknown.New("PGDate: unknown time error")
}

// Value implements the driver Valuer interface.
func (d PGDate) Value() (driver.Value, error) {
	return d.String(), nil
}

// MarshalJSON for PGDate
func (d PGDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON for PGDate
func (d *PGDate) UnmarshalJSON(data []byte) error {
	var da string
	if err := json.Unmarshal(data, &da); err != nil {
		return err
	}

	if err := d.Set(da); err != nil {
		return err
	}

	return nil
}

func (d PGDate) CurrentDateOrAfter(t time.Time) bool {
	return d.After(t) || (t.YearDay() == d.YearDay())
}
