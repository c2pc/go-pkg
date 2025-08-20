package clause

import (
	"reflect"
	"strings"
)

func GetNotNullDBFields(s any) []any {
	val := reflect.ValueOf(s)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typeOfS := val.Type()
	numFields := typeOfS.NumField()
	result := make([]any, 0, numFields)

	for i := 0; i < numFields; i++ {
		field := typeOfS.Field(i)
		fieldValue := val.Field(i)

		if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
			dbTag := field.Tag.Get("db")

			dbFieldName := strings.Split(dbTag, ",")[0]
			if dbFieldName != "" {
				result = append(result, dbFieldName)
			}
		}
	}

	return result
}

func PtrToValue[T comparable](ptr *T) (defaultValue T) {
	if ptr != nil {
		return *ptr
	}

	return defaultValue
}

func PtrToNullableValue[T comparable](ptr *T) *T {
	var defaultValue T

	if ptr != nil {
		if *ptr == defaultValue {
			return nil
		}
		return ptr
	}

	return nil
}
