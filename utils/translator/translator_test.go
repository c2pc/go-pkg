package translator

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetValidateTranslators(t *testing.T) {
	validate := validator.New()
	SetValidateTranslators(validate)

	// Ensure that the default Russian translations are correctly registered
	_, found := utrans.GetTranslator("ru")
	if !found {
		assert.Fail(t, "translator not found")
	}
}

func TestGetTranslator(t *testing.T) {
	validate := validator.New()
	SetValidateTranslators(validate)

	tests := []struct {
		name       string
		acceptLang string
		wantLang   string
	}{
		{
			name:       "Get English Translator",
			acceptLang: "en",
			wantLang:   "en",
		},
		{
			name:       "Get Russian Translator",
			acceptLang: "ru",
			wantLang:   "ru",
		},
		{
			name:       "Get Default Translator",
			acceptLang: "fr",
			wantLang:   "ru",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTranslator(tt.acceptLang)
			lang := got.Locale()
			assert.Equal(t, tt.wantLang, lang, "should return the correct translator locale")
		})
	}
}

func TestRegisterValidatorTranslation(t *testing.T) {
	validate := validator.New()
	SetValidateTranslators(validate)

	_, trans, addFunc, translateFunc := RegisterValidatorTranslation(
		"ru",
		"required",
		"Поле обязательно для заполнения",
		true,
	)

	// Register the custom translation
	err := addFunc(trans)
	if err != nil {
		t.Fatalf("error registering custom translation: %v", err)
	}

	// Validate an empty field to trigger the "required" validation error
	type TestStruct struct {
		Name string `validate:"required"`
	}

	testStruct := TestStruct{}
	errs := validate.Struct(testStruct)
	if errs == nil {
		t.Fatalf("expected validation error, got nil")
	}

	// Convert the validation error to a FieldError
	var fieldErr validator.FieldError
	ok := errors.As(errs.(validator.ValidationErrors)[0], &fieldErr)
	if !ok {
		t.Fatalf("failed to convert error to FieldError")
	}

	// Translate the error
	result := translateFunc(trans, fieldErr)
	expected := "Поле обязательно для заполнения"
	assert.Equal(t, expected, result, "should return the correct translated error message")
}
