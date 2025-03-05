package validator

import (
	"fmt"
	"reflect"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/go-playground/validator/v10"
)

var DeviceID validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()

	if kind == reflect.Int || kind == reflect.Int64 {
		deviceID := field.Int()
		return model.DeviceIDToName(int(deviceID)) != ""
	} else {
		panic(fmt.Sprintf("Bad type for %s", fl.FieldName()))
	}
}

func DeviceIDValidation(v *validator.Validate) {
	_ = v.RegisterValidation("device_id", DeviceID, false)
	_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.RU, "device_id", "{0} неизвестное устройство", true))
	_ = v.RegisterTranslation(translator.RegisterValidatorTranslation(translator.EN, "device_id", "{0] unknown device", true))
}
