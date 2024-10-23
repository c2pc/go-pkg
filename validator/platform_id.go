package validator

import (
	"fmt"
	"reflect"

	"github.com/c2pc/go-pkg/platform"
	"github.com/go-playground/validator/v10"
)

var ValidatePlatformID validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()

	if kind == reflect.Int || kind == reflect.Int64 {
		platformID := field.Int()
		return platform.PlatformIDToName(int(platformID)) != ""
	} else {
		panic(fmt.Sprintf("Bad type for %s", fl.FieldName()))
	}
}
