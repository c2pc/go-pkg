package translator

import (
	"testing"
)

func TestTranslate(t *testing.T) {
	tests := []struct {
		name       string
		transMap   Translate
		acceptLang string
		args       []interface{}
		want       string
	}{
		{
			name: "translation exists for acceptLang",
			transMap: Translate{
				EN: "Hello, World!",
				RU: "Привет, мир!",
			},
			acceptLang: "en",
			args:       nil,
			want:       "Hello, World!",
		},
		{
			name: "translation does not exist for acceptLang, but exists for RU",
			transMap: Translate{
				RU: "Привет, мир!",
			},
			acceptLang: "en",
			args:       nil,
			want:       "Привет, мир!",
		},
		{
			name: "no translation exists",
			transMap: Translate{
				EN: "Hello, World!",
			},
			acceptLang: "fr",
			args:       nil,
			want:       "",
		},
		{
			name: "translation exists and args are provided",
			transMap: Translate{
				EN: "Hello, %s!",
				RU: "Привет, %s!",
			},
			acceptLang: "en",
			args:       []interface{}{"John"},
			want:       "Hello, John!",
		},
		{
			name: "translation does not exist for acceptLang, RU used with args",
			transMap: Translate{
				RU: "Привет, %s!",
			},
			acceptLang: "fr",
			args:       []interface{}{"Иван"},
			want:       "Привет, Иван!",
		},
		{
			name:       "no translations and args provided",
			transMap:   Translate{},
			acceptLang: "fr",
			args:       []interface{}{"Иван"},
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.transMap.Translate(tt.acceptLang, tt.args...)
			if got != tt.want {
				t.Errorf("Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}
