package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/go-playground/validator/v10"
)

var SpecChars validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()

	if kind == reflect.String {
		s := field.String()

		isValid := regexp.MustCompile("^[\\sa-zA-Z0-9а-яА-ЯёЁ`~!@#$%^&*()_+={}\\[\\]\\\\|:;\"/'<>,.?-]*$").MatchString
		noSpaces := !strings.HasPrefix(s, " ") && !strings.HasSuffix(s, " ")

		return isValid(s) && noSpaces
	} else {
		panic(fmt.Sprintf("Bad type for %s", fl.FieldName()))
	}
}

func SpecCharsValidation(v *validator.Validate) {
	_ = v.RegisterValidation("spec_chars", SpecChars, false)
	_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.RU, "spec_chars", "{0} должен содержать только символы \\sa-zA-Z0-9а-яА-ЯёЁ`~!@#$%^&*()_+={}\\[\\]\\\\|:;\"/'<>,.?-", true))
	_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.EN, "spec_chars", "{0} must contain only characters \\sa-zA-Z0-9а-яА-ЯёЁ`~!@#$%^&*()_+={}\\[\\]\\\\|:;\"/'<>,.?-", true))
}
