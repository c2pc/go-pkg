package jsonutil

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: If you also want tests that include protobuf messages, define a.proto file first, and then use the protoc command to generate the Go code

func TestJsonMarshal(t *testing.T) {
	structData := struct{ Name string }{"John"}
	structBytes, err := JsonMarshal(structData)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"Name":"John"}`, string(structBytes))

	marshalerData := json.RawMessage(`{"type":"raw"}`)
	marshalerBytes, err := JsonMarshal(marshalerData)
	assert.NoError(t, err)
	assert.Equal(t, `{"type":"raw"}`, string(marshalerBytes))
}

func TestJsonUnmarshal(t *testing.T) {
	structBytes := []byte(`{"Name":"Jane"}`)
	var structData struct{ Name string }
	err := JsonUnmarshal(structBytes, &structData)
	assert.NoError(t, err)
	assert.Equal(t, "Jane", structData.Name)

	marshalerBytes := []byte(`{"type":"unmarshal"}`)
	var marshalerData json.RawMessage
	err = JsonUnmarshal(marshalerBytes, &marshalerData)
	assert.NoError(t, err)
	assert.Equal(t, `{"type":"unmarshal"}`, string(marshalerData))
}
