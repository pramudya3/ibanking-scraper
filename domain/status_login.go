package domain

import (
	"database/sql/driver"
	"encoding/json"
)

const (
	StatusUnknown LoginStatus = iota
	BelumLogin
	ProsesLogin
	SudahLogin
	ProsesScraping
	ButuhToken
	TidakAktif
)

type LoginStatus uint

func (c LoginStatus) Value() (driver.Value, error) {
	return c.String(), nil
}

func (c *LoginStatus) Scan(value interface{}) error {
	if val, ok := value.(string); ok {
		*c = c.getCode(val)
	}
	return nil
}

func (c LoginStatus) MarshalJSON() ([]byte, error) {
	if c == StatusUnknown {
		return json.Marshal(nil)
	}
	return json.Marshal(c.String())
}

func (c *LoginStatus) UnmarshalJSON(data []byte) error {
	var t string
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	*c = c.getCode(t)
	return nil
}

func (c LoginStatus) String() string {
	switch c {
	case BelumLogin:
		return "Belum Login"
	case ProsesLogin:
		return "Proses Login"
	case SudahLogin:
		return "Sudah Login"
	case ProsesScraping:
		return "Proses Scraping"
	case ButuhToken:
		return "Butuh Token"
	case TidakAktif:
		return "Tidak Aktif"
	default:
		return "Status Unknown"
	}
}

func (c *LoginStatus) getCode(val string) LoginStatus {
	switch val {
	case "Belum Login":
		return BelumLogin
	case "Proses Login":
		return ProsesLogin
	case "Sudah Login":
		return SudahLogin
	case "Proses Scraping":
		return ProsesScraping
	case "Butuh Token":
		return ButuhToken
	case "Tidak Aktif":
		return TidakAktif
	default:
		return StatusUnknown
	}
}
