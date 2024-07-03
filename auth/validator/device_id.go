package validator

import (
	"fmt"
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/go-playground/validator/v10"
	"reflect"
)

var ValidateDeviceID validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()

	if kind == reflect.Int || kind == reflect.Int64 {
		deviceID := field.Int()
		return model.DeviceIDToName(int(deviceID)) != ""
	} else {
		panic(fmt.Sprintf("Bad type for %s", fl.FieldName()))
	}
}
