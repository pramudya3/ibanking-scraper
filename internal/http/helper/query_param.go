package helper

import (
	"fmt"
	"ibanking-scraper/internal/validator/v1"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func ReadInt(query url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := query.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddFieldError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

func ReadString(query url.Values, key string, defaultValue string) string {
	s := query.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func ReadStringWithValidator(query url.Values, key string, defaultValue string, v *validator.Validator, fn func(string) bool) string {
	s := query.Get(key)

	if s == "" {
		return defaultValue
	}

	v.Check(fn(s), key, fmt.Sprintf("unknown value for '%s", s))
	return s
}

func ReadStringWithRegex(query url.Values, key string, defaultValue string, v *validator.Validator, rx *regexp.Regexp) string {
	s := query.Get(key)

	if s == "" {
		return defaultValue
	}

	v.Check(validator.Matches(s, rx), key, fmt.Sprintf("unknown value for '%s'", s))
	return s
}

func ReadCSV(query url.Values, key string, defaultValue []string) []string {
	csv := query.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}
