package validator

import "github.com/go-playground/validator/v10"

var rules = map[string]string{
	"email":    "please provide valid email",
	"required": "required field is empty",
}

func getErrorMessage(err validator.FieldError) string {
	// Field tag: err.Tag()
	// Field name: err.Field()
	// Field value: err.Value()

	if message, ok := rules[err.Tag()]; ok {
		return message
	}

	return "value is not acceptable"
}
