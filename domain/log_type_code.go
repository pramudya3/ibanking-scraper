package domain

import (
	"database/sql/driver"
	"encoding/json"
)

const (
	TypeLog LogType = iota
	TypeError
	TypeInfo
	TypeUnknown
)

type LogType uint

func (c LogType) Value() (driver.Value, error) {
	return c.String(), nil
}

func (c *LogType) Scan(value interface{}) error {
	if val, ok := value.(string); ok {
		*c = c.getCode(val)
	}

	return nil
}

func (c LogType) MarshalJSON() ([]byte, error) {
	if c == TypeUnknown {
		return json.Marshal(nil)
	}
	return json.Marshal(c.String())
}

func (c *LogType) UnmarshalJSON(data []byte) error {
	var t string
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	*c = c.getCode(t)
	return nil
}

func (c LogType) String() string {
	switch c {
	case TypeLog:
		return "log"
	case TypeError:
		return "error"
	case TypeInfo:
		return "info"
	default:
		return "Unknown"
	}
}

func (c *LogType) getCode(val string) LogType {
	switch val {
	case "log":
		return TypeLog
	case "error":
		return TypeError
	case "info":
		return TypeInfo
	default:
		return TypeUnknown
	}
}
