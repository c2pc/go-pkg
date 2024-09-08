package jsonutil

import (
	"encoding/json"
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
