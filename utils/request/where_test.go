package request

import (
	"errors"
	"testing"

	"github.com/c2pc/go-pkg/v2/utils/clause"
)

func TestParseWhere(t *testing.T) {
	tests := []struct {
		input       string
		expected    *clause.ExpressionWhere
		expectedErr error
	}{
		{
			input:    "",
			expected: nil,
		},
		{
			input: "a = 1",
			expected: &clause.ExpressionWhere{
				Expressions: nil,
				Column:      "a",
				Operation:   "=",
				Value:       "1",
			},
		},
		{
			input: "a in [ a , b,c]",
			expected: &clause.ExpressionWhere{
				Expressions: nil,
				Column:      "a",
				Operation:   "in",
				Value:       "a,b,c",
			},
		},
		{
			input: "a in [`a`,`b`,`c`]",
			expected: &clause.ExpressionWhere{
				Expressions: nil,
				Column:      "a",
				Operation:   "in",
				Value:       "`a`,`b`,`c`",
			},
		},
		{
			input: "a in [ `a`]",
			expected: &clause.ExpressionWhere{
				Expressions: nil,
				Column:      "a",
				Operation:   "in",
				Value:       "`a`",
			},
		},
		{
			input: "a in []",
			expected: &clause.ExpressionWhere{
				Expressions: nil,
				Column:      "a",
				Operation:   "in",
				Value:       "",
			},
		},
		{
			input: "a = 1 and c = ` qwerty ` ",
			expected: &clause.ExpressionWhere{
				Expressions: []clause.ExpressionWhere{
					{Column: "a", Operation: "=", Value: "1"},
					{Operation: "and"},
					{Column: "c", Operation: "=", Value: "` qwerty `"},
				},
			},
		},
		{
			input: "( a = 1 )",
			expected: &clause.ExpressionWhere{
				Expressions: []clause.ExpressionWhere{
					{Column: "a", Operation: "=", Value: "1"},
				},
			},
		},
		{
			input: "a = 1 and b co `test`",
			expected: &clause.ExpressionWhere{
				Expressions: []clause.ExpressionWhere{
					{Column: "a", Operation: "=", Value: "1"},
					{Operation: "and"},
					{Column: "b", Operation: "co", Value: "`test`"},
				},
			},
		},
		{
			input: "a pt and b np",
			expected: &clause.ExpressionWhere{
				Expressions: []clause.ExpressionWhere{
					{Column: "a", Operation: "pt", Value: ""},
					{Operation: "and"},
					{Column: "b", Operation: "np", Value: ""},
				},
			},
		},
		{
			input: "a in [true,false] and b eq true or c nin [false]",
			expected: &clause.ExpressionWhere{
				Expressions: []clause.ExpressionWhere{
					{Column: "a", Operation: "in", Value: "true,false"},
					{Operation: "and"},
					{Column: "b", Operation: "eq", Value: "true"},
					{Operation: "or"},
					{Column: "c", Operation: "nin", Value: "false"},
				},
			},
		},
		{
			input: "a pt",
			expected: &clause.ExpressionWhere{
				Expressions: nil,
				Column:      "a",
				Operation:   "pt",
				Value:       "",
			},
		},
		{
			input: "a = 1 and ( b co `test` )",
			expected: &clause.ExpressionWhere{
				Expressions: []clause.ExpressionWhere{
					{Column: "a", Operation: "=", Value: "1"},
					{Operation: "and"},
					{Column: "b", Operation: "co", Value: "`test`"},
				},
			},
		},
		{
			input: "a = 1 and ( b co `test` or c = 10 )",
			expected: &clause.ExpressionWhere{
				Expressions: []clause.ExpressionWhere{
					{Column: "a", Operation: "=", Value: "1"},
					{Operation: "and"},
					{
						Expressions: []clause.ExpressionWhere{
							{Column: "b", Operation: "co", Value: "`test`"},
							{Operation: "or"},
							{Column: "c", Operation: "=", Value: "10"},
						},
					},
				},
			},
		},
		{
			input: "( a = 1 and ( b co `test` or c = 10 ) )",
			expected: &clause.ExpressionWhere{
				Expressions: []clause.ExpressionWhere{
					{
						Expressions: []clause.ExpressionWhere{
							{Column: "a", Operation: "=", Value: "1"},
							{Operation: "and"},
							{
								Expressions: []clause.ExpressionWhere{
									{Column: "b", Operation: "co", Value: "`test`"},
									{Operation: "or"},
									{Column: "c", Operation: "=", Value: "10"},
								},
							},
						},
					},
				},
			},
		},
		{
			input: "a = 1 and ( b co `test` or c = 10 ) or (d = 20 and (e > 30 and f < 40 and g pt) or h np and i >= 20)",
			expected: &clause.ExpressionWhere{
				Expressions: []clause.ExpressionWhere{
					{Column: "a", Operation: "=", Value: "1"},
					{Operation: "and"},
					{
						Expressions: []clause.ExpressionWhere{
							{Column: "b", Operation: "co", Value: "`test`"},
							{Operation: "or"},
							{Column: "c", Operation: "=", Value: "10"},
						},
					},
					{Operation: "or"},
					{
						Expressions: []clause.ExpressionWhere{
							{Column: "d", Operation: "=", Value: "20"},
							{Operation: "and"},
							{
								Expressions: []clause.ExpressionWhere{
									{Column: "e", Operation: ">", Value: "30"},
									{Operation: "and"},
									{Column: "f", Operation: "<", Value: "40"},
									{Operation: "and"},
									{Column: "g", Operation: "pt", Value: ""},
								},
							},
							{Operation: "or"},
							{Column: "h", Operation: "np", Value: ""},
							{Operation: "and"},
							{Column: "i", Operation: ">=", Value: "20"},
						},
					},
				},
			},
		},
		{
			input: "a eq true and b = false or c = false",
			expected: &clause.ExpressionWhere{
				Expressions: []clause.ExpressionWhere{
					{Column: "a", Operation: "eq", Value: "true"},
					{Operation: "and"},
					{Column: "b", Operation: "=", Value: "false"},
					{Operation: "or"},
					{Column: "c", Operation: "=", Value: "false"},
				},
			},
		},
		{
			input:       "a eq",
			expected:    nil,
			expectedErr: errors.New("syntax_error.(invalid tokens: [a eq])"),
		},
		{
			input:       "a noop `qwer`",
			expected:    nil,
			expectedErr: errors.New("syntax_error.(invalid operator: noop)"),
		},
		{
			input:       "a = `qwer` or ( b noop `test` )",
			expected:    nil,
			expectedErr: errors.New("syntax_error.(invalid operator: noop)"),
		},
		{
			input:       "a = 1 and (a = 20",
			expectedErr: errors.New("syntax_error.(invalid input sequence)"),
		},
		{
			input:       "a = 1 and (a = 20]",
			expectedErr: errors.New("syntax_error.(invalid input sequence)"),
		},
		{
			input: "a = 1 and a in [a = 10 and b = 20]",
			expected: &clause.ExpressionWhere{
				Expressions: []clause.ExpressionWhere{
					{Column: "a", Operation: "=", Value: "1"},
					{Operation: "and"},
					{Column: "a", Operation: "in", Value: "a,=,10,and,b,=,20"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := ParseWhere(test.input)
			if err != nil {
				if test.expectedErr == nil {
					t.Fatalf("unexpected error: %v", err)
				} else if err.Error() != test.expectedErr.Error() {
					t.Fatalf("expected error message %q but got %q", test.expectedErr.Error(), err.Error())
				}
			} else {
				if test.expectedErr != nil {
					t.Fatalf("expected error message %q but got nil", test.expectedErr.Error())
				}
			}

			if !compareExpressions(result, test.expected) {
				t.Errorf("For input %v, \nexpected %+v, \nbut got %+v", test.input, test.expected, result)
			}
		})
	}
}

// Сравнение двух выражений
func compareExpressions(a, b *clause.ExpressionWhere) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if a.Column != b.Column {
		return false
	}
	if a.Operation != b.Operation {
		return false
	}
	if a.Value != b.Value {
		return false
	}
	if len(a.Expressions) != len(b.Expressions) {
		return false
	}
	for i := range a.Expressions {
		if !compareExpressions(&a.Expressions[i], &b.Expressions[i]) {
			return false
		}
	}
	return true
}
