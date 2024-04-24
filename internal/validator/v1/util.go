package validator

import (
	"ibanking-scraper/internal/tool"
	"ibanking-scraper/pkg/constant"
	"regexp"
	"time"
)

var PhoneNumberRX = regexp.MustCompile(`^[0-9]{10,14}$`)

// In returns true if a specific value is in a list of strings.
func In(value string, list ...string) bool {
	return tool.StringInArr(value, list...)
}

func InUint(value uint64, list ...uint64) bool {
	return tool.Uint64InArr(value, list...)
}

// Matches returns true if a string value matches a specific regexp pattern.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique returns true if all string values in a slice are unique.
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}

// Date returns true if value is compatible with layout ISO8601/YYYY-MM-DD.
func Date(value string) bool {
	if _, err := time.Parse(constant.LayoutDateISO8601, value); err != nil {
		return false
	}

	return true
}

// Time returns true if value is compatible with layout Time/ HH:mm:ss.
func Time(value string) bool {
	if _, err := time.Parse(constant.LayoutTime, value); err != nil {
		return false
	}

	return true
}

func TimeWithoutSecond(value string) bool {
	if _, err := time.Parse(constant.LayoutTimeWithoutSecond, value); err != nil {
		return false
	}

	return true
}
