package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/go-playground/validator/v10"
)

var ExcludeSpace validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()

	if kind == reflect.String {
		s := field.String()

		return !strings.Contains(s, " ")

	} else {
		panic(fmt.Sprintf("Bad type for %s", fl.FieldName()))
	}
}

func ExcludeSpaceValidation(v *validator.Validate) {
	_ = v.RegisterValidation("exclude_space", ExcludeSpace, false)
	_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.RU, "exclude_space", "{0} не должен содержать пробел", true))
	_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.EN, "exclude_space", "{0} must not contain spaces", true))
}
