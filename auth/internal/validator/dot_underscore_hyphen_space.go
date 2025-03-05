package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/go-playground/validator/v10"
)

var DotUnderscoreHyphenSpace validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()

	if kind == reflect.String {
		s := field.String()

		isValid := regexp.MustCompile(`^[\sa-zA-Z0-9а-яА-ЯёЁ_.-]*$`).MatchString
		noSpaces := !strings.HasPrefix(s, " ") && !strings.HasSuffix(s, " ")

		return isValid(s) && noSpaces
	} else {
		panic(fmt.Sprintf("Bad type for %s", fl.FieldName()))
	}
}

func DotUnderscoreHyphenSpaceValidation(v *validator.Validate) {
	_ = v.RegisterValidation("dot_underscore_hyphen_space", DotUnderscoreHyphenSpace, false)
	_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.RU, "dot_underscore_hyphen_space", "{0} должен содержать только символы \\sa-zA-Z0-9а-яА-ЯёЁ_.- и не должен начинаться или заканчиваться пробелом", true))
	_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.EN, "dot_underscore_hyphen_space", "{0} must contain only characters \\sa-zA-Z0-9а-яА-ЯёЁ_.- and must not begin or end with a space", true))
}
