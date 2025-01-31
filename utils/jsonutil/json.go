package jsonutil

import (
	"encoding/json"
	"strings"
)

func JsonMarshal(v any) ([]byte, error) {
	m, err := json.Marshal(v)
	return m, err
}

func JsonUnmarshal(b []byte, v any) error {
	return json.Unmarshal(b, v)
}

func StructToJsonString(param any) string {
	dataType, _ := JsonMarshal(param)
	dataString := string(dataType)
	return dataString
}

// JsonStringToStruct The incoming parameter must be a pointer
func JsonStringToStruct(s string, args any) error {
	err := json.Unmarshal([]byte(s), args)
	return err
}

func JsonHideImportantData(input []byte, keys ...string) []byte {
	if len(keys) == 0 {
		return input
	}

	if input == nil {
		return input
	}

	var data map[string]interface{}
	if err := json.Unmarshal(input, &data); err != nil {
		return input
	}

	maskSensitiveFields(data, keys...)

	output, err := json.Marshal(data)
	if err != nil {
		return input
	}

	return output
}

func maskSensitiveFields(data map[string]interface{}, keys ...string) {
	for key, value := range data {
		for _, sensitiveKey := range keys {
			if strings.Contains(strings.ToLower(key), strings.ToLower(sensitiveKey)) {
				if _, ok := value.(string); ok {
					data[key] = "****"
					break
				}
			}
		}

		if nested, ok := value.(map[string]interface{}); ok {
			maskSensitiveFields(nested, keys...)
		} else if array, ok := value.([]interface{}); ok {
			for _, item := range array {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					maskSensitiveFields(nestedMap, keys...)
				}
			}
		}
	}
}
