package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/go-playground/validator/v10"
)

var PhoneNumber validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()

	if kind == reflect.String {
		s := field.String()

		if len(s) == 0 {
			return true
		}

		isValid := regexp.MustCompile(`^[\s0-9+*()-]*$`).MatchString
		noSpaces := !strings.HasPrefix(s, " ") && !strings.HasSuffix(s, " ")

		return isValid(s) && noSpaces
	} else {
		panic(fmt.Sprintf("Bad type for %s", fl.FieldName()))
	}
}

func PhoneNumberValidation(v *validator.Validate) {
	_ = v.RegisterValidation("phone_number", PhoneNumber, false)
	_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.RU, "phone_number", "{0} должен содержать только символы \\s0-9+*()- и не должен начинаться или заканчиваться пробелом", true))
	_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.EN, "phone_number", "{0} must contain only characters \\s0-9+*()- and must not begin or end with a space", true))
}
