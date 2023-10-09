package apperr

import (
	"testing"
)

func TestGetNamespace(t *testing.T) {
	cases := []struct {
		name   string
		output string
	}{
		{
			name:   "name",
			output: "name",
		},
		{
			name:   "user_id",
			output: "user_id",
		},
		{
			name:   "User.id",
			output: "id",
		},
		{
			name:   "Settings.name",
			output: "name",
		},
		{
			name:   "User.Numbers[0].name",
			output: "numbers[0].name",
		},
		{
			name:   "User.Numbers[0].Number.name",
			output: "numbers[0].number.name",
		},
		{
			name:   "User.Profiles[1].Numbers[0].Number.name",
			output: "profiles[1].numbers[0].number.name",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			output := getNamespace(tt.name)
			if output != tt.output {
				t.Errorf("%s <> %s", output, tt.output)
			}
		})
	}
}
