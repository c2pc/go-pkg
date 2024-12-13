package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var DotUnderscoreHyphenSpace validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()

	if kind == reflect.String {
		s := field.String()

		isValid := regexp.MustCompile(`^[\sa-zA-Z0-9а-яА-Я_.-]*$`).MatchString
		noSpaces := !strings.HasPrefix(s, " ") && !strings.HasSuffix(s, " ")

		return isValid(s) && noSpaces
	} else {
		panic(fmt.Sprintf("Bad type for %s", fl.FieldName()))
	}
}
