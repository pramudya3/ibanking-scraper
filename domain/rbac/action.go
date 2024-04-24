package rbac

import (
	"database/sql/driver"
	"encoding/json"
	"ibanking-scraper/internal/tool"
	"net/http"
)

type Action uint

const (
	Edit Action = iota + 1
	Lihat
	Blok
)

func (a Action) String() string {
	return []string{
		"unknown",
		"edit",
		"lihat",
		"blok",
	}[a]
}

func (a *Action) fromString(act string) {
	switch act {
	case "edit":
		*a = Edit
	case "lihat":
		*a = Lihat
	case "blok":
		*a = Blok
	}
}

func (a Action) FromHTTPMethod(httpMethod string) bool {
	switch a {
	case Edit:
		return true
	case Lihat:
		return tool.StringInArr(httpMethod, http.MethodGet)
	case Blok:
		return false
	}

	return false
}

func (a Action) Value() (driver.Value, error) {
	if a == 0 {
		return nil, nil
	}

	return a.String(), nil
}

func (a *Action) Scan(value interface{}) error {
	val, ok := value.(string)
	if !ok {
		*a = Action(0)
	}

	a.fromString(val)

	return nil
}

func (a Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *Action) UnmarshalJSON(data []byte) error {
	var act string
	if err := json.Unmarshal(data, &act); err != nil {
		return err
	}

	a.fromString(act)
	return nil
}
