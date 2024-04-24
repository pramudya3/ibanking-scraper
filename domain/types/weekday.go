package types

import (
	"database/sql/driver"
	"encoding/json"
	"ibanking-scraper/internal/errors"
)

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	UnknownDay
)

type Weekday int

func (w Weekday) fromString(str string) Weekday {
	switch str {
	case "sun":
		return Sunday
	case "mon":
		return Monday
	case "tue":
		return Tuesday
	case "wed":
		return Wednesday
	case "thu":
		return Thursday
	case "fri":
		return Friday
	case "sat":
		return Saturday
	}

	return UnknownDay
}

func (w Weekday) String() string {
	switch w {
	case Sunday:
		return "sun"
	case Monday:
		return "mon"
	case Tuesday:
		return "tue"
	case Wednesday:
		return "wed"
	case Thursday:
		return "thu"
	case Friday:
		return "fri"
	case Saturday:
		return "sat"
	}

	return "sun"
}

// Scan implements the Scanner interface for Weekday
func (w *Weekday) Scan(value interface{}) error {
	if v, ok := value.(string); ok {
		*w = w.fromString(v)
		return nil
	}

	return errors.Unknown.New("unexpected value for weekday")
}

// Value implements the driver Valuer interface.
func (w Weekday) Value() (driver.Value, error) {
	return w.String(), nil
}

// MarshalJSON for Weekday
func (w Weekday) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.String())
}

// UnmarshalJSON for Weekday
func (w *Weekday) UnmarshalJSON(data []byte) error {
	var t string
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	*w = w.fromString(t)

	return nil
}

type Days []Weekday

func (d Days) In(day Weekday) bool {
	for _, da := range d {
		if da == day {
			return true
		}
	}

	return false
}

func NewDaysFromStrings(days []string) Days {
	d := make(Days, len(days))

	for i, day := range days {
		d[i] = new(Weekday).fromString(day)
	}

	return d
}
