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

func JsonClearPassword(input []byte) []byte {
	var data map[string]interface{}
	if err := json.Unmarshal(input, &data); err != nil {
		return input
	}

	maskSensitiveFields(data)

	output, err := json.Marshal(data)
	if err != nil {
		return input
	}
	return output
}

func maskSensitiveFields(data map[string]interface{}) {
	for key, value := range data {
		if strings.Contains(strings.ToLower(key), "pass") || strings.Contains(strings.ToLower(key), "pwd") {
			if _, ok := value.(string); ok {
				data[key] = "****"
			}
		} else if nested, ok := value.(map[string]interface{}); ok {
			maskSensitiveFields(nested)
		} else if array, ok := value.([]interface{}); ok {
			for _, item := range array {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					maskSensitiveFields(nestedMap)
				}
			}
		}
	}
}
