package validator

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
)

func ValidateRegex(regexp *regexp.Regexp) func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		field := fl.Field()
		kind := field.Kind()

		if kind == reflect.Ptr {
			if field.IsNil() {
				return false
			}
			field = field.Elem()
			kind = field.Kind()
		}

		if kind != reflect.String {
			return false
		}

		value := field.String()

		isValid := regexp.MatchString

		return isValid(value)
	}
}
