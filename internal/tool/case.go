package tool

import (
	"fmt"
	"regexp"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func ArrayToString(val interface{}, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(val), " ", delim, -1), "[]")
}

// StringInArr returns true if a specific value is in a list of strings.
func StringInArr(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// Int64InArr returns true if a specific value is in a list of uint.
func Int64InArr(value int64, list ...int64) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// Uint64InArr returns true if a specific value is in a list of uint.
func Uint64InArr(value uint64, list ...uint64) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// StringToSlug return slug of text
func StringToSlug(text string) string {
	regex := regexp.MustCompile(`[^\d\w ]+`)
	cleanText := regex.ReplaceAllString(text, "")
	lowerText := strings.ToLower(cleanText)
	slug := strings.ReplaceAll(lowerText, " ", "-")
	return slug
}
