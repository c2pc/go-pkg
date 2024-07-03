package validator

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

var ValidateEqualsIf validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()

	params := ParseOneOfParam2(fl.Param())
	if len(params) < 3 || (len(params)-1)%2 != 0 {
		panic(fmt.Sprintf("Bad param number for equals_if %s", fl.FieldName()))
	}

	for i := 1; i < len(params); i += 2 {
		if !RequireCheckFieldValue(fl, params[i], params[i+1], false) {
			return true
		}
	}

	return CheckField(field, kind, params[0])
}
