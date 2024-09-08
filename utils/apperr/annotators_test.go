package apperr

import (
	translator2 "github.com/c2pc/go-pkg/v2/utils/translator/mock"
	"testing"

	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
)

func TestAnnotatorsWithID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "set non-empty ID",
			id:   "12345",
			want: "12345",
		},
		{
			name: "empty ID",
			id:   "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{}
			WithID(tt.id)(err)
			if err.ID != tt.want {
				t.Errorf("expected ID %v, got %v", tt.want, err.ID)
			}
		})
	}
}

func TestAnnotatorsWithTextTranslate(t *testing.T) {
	mockTranslate := translator2.MockTranslator{}

	tests := []struct {
		name string
		tr   translator.Translator
		want translator.Translator
	}{
		{
			name: "set non-nil translator",
			tr:   &mockTranslate,
			want: &mockTranslate,
		},
		{
			name: "nil translator",
			tr:   nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{}
			WithTextTranslate(tt.tr)(err)
			if err.TextTranslate != tt.want {
				t.Errorf("expected translator %v, got %v", tt.want, err.TextTranslate)
			}
		})
	}
}

func TestAnnotatorsWithCode(t *testing.T) {
	tests := []struct {
		name string
		code code.Code
		want code.Code
	}{
		{
			name: "set code",
			code: code.NotFound,
			want: code.NotFound,
		},
		{
			name: "set different code",
			code: code.Internal,
			want: code.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{}
			WithCode(tt.code)(err)
			if err.Code != tt.want {
				t.Errorf("expected code %v, got %v", tt.want, err.Code)
			}
		})
	}
}

func TestAnnotatorsWithText(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "set non-empty text",
			text: "error occurred",
			want: "error occurred",
		},
		{
			name: "empty text",
			text: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{}
			WithText(tt.text)(err)
			if err.Text != tt.want {
				t.Errorf("expected text %v, got %v", tt.want, err.Text)
			}
		})
	}
}
