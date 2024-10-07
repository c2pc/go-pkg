package validator

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	oneofValsCache       = map[string][]string{}
	oneofValsCacheRWLock = sync.RWMutex{}
)

func ParseOneOfParam2(s string) []string {
	oneofValsCacheRWLock.RLock()
	vals, ok := oneofValsCache[s]
	oneofValsCacheRWLock.RUnlock()
	if !ok {
		oneofValsCacheRWLock.Lock()
		vals = regexp.MustCompile(`'[^']*'|\S+`).FindAllString(s, -1)
		for i := 0; i < len(vals); i++ {
			vals[i] = strings.Replace(vals[i], "'", "", -1)
		}
		oneofValsCache[s] = vals
		oneofValsCacheRWLock.Unlock()
	}
	return vals
}

// HasValue is the validation function for validating if the current field's value is not the default static value.
func HasValue(fl validator.FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !field.IsNil()
	default:
		if field.Interface() != nil {
			return true
		}
		return field.IsValid() && field.Interface() != reflect.Zero(field.Type()).Interface()
	}
}

func RequireCheckFieldValue(fl validator.FieldLevel, param string, value string, defaultNotFoundValue bool) bool {
	field, kind, _, found := fl.GetStructFieldOKAdvanced2(fl.Parent(), param)
	if !found {
		return defaultNotFoundValue
	}

	return CheckField(field, kind, value)
}

func CheckField(field reflect.Value, kind reflect.Kind, value string) bool {
	switch kind {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == AsInt(value)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return field.Uint() == AsUint(value)

	case reflect.Float32, reflect.Float64:
		return field.Float() == AsFloat(value)

	case reflect.Slice, reflect.Map, reflect.Array:
		return int64(field.Len()) == AsInt(value)

	case reflect.Bool:
		return field.Bool() == AsBool(value)
	default:
		return field.String() == value
	}
}

// AsInt returns the parameter as a int64
// or panics if it can't convert
func AsInt(param string) int64 {
	i, err := strconv.ParseInt(param, 0, 64)
	panicIf(err)

	return i
}

// AsUint returns the parameter as a uint64
// or panics if it can't convert
func AsUint(param string) uint64 {

	i, err := strconv.ParseUint(param, 0, 64)
	panicIf(err)

	return i
}

// AsFloat returns the parameter as a float64
// or panics if it can't convert
func AsFloat(param string) float64 {

	i, err := strconv.ParseFloat(param, 64)
	panicIf(err)

	return i
}

// AsBool returns the parameter as a bool
// or panics if it can't convert
func AsBool(param string) bool {

	i, err := strconv.ParseBool(param)
	panicIf(err)

	return i
}

func panicIf(err error) {
	if err != nil {
		panic(err.Error())
	}
}
