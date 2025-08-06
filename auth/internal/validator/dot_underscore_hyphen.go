package validator

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/go-playground/validator/v10"
)

var DotUnderscoreHyphen validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()

	if kind == reflect.String {
		s := field.String()

		isValid := regexp.MustCompile(`^[a-zA-Z0-9а-яА-ЯёЁ_.!@-]*$`).MatchString

		return isValid(s)
	} else {
		panic(fmt.Sprintf("Bad type for %s", fl.FieldName()))
	}
}

func DotUnderscoreHyphenValidation(v *validator.Validate) {
	_ = v.RegisterValidation("dot_underscore_hyphen", DotUnderscoreHyphen, false)
	_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.RU, "dot_underscore_hyphen", "{0} должен содержать только символы a-zA-Z0-9а-яА-ЯёЁ_.!@-", true))
	_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.EN, "dot_underscore_hyphen", "{0} must contain only characters a-zA-Z0-9а-яА-ЯёЁ_.!@-", true))
}
