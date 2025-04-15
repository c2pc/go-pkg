package clause

import (
	"strings"
	"testing"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/stretchr/testify/assert"
)

var quoteTo = func(s string) string { return strings.ReplaceAll(s, `"`, "`") }
var emptyJoins []string
var emptyArgs []interface{}

func TestWhereFilter_AllCases(t *testing.T) {
	fieldSearchable := FieldSearchable{
		"name":   {Column: "user_name", Type: String},
		"age":    {Column: "user_age", Type: Int},
		"active": {Column: "is_active", Type: Bool},
		"date":   {Column: "created_at", Type: DateTime},
	}

	// Nested expression test
	t.Run("nested expressions", func(t *testing.T) {
		expr := &ExpressionWhere{
			Expressions: []ExpressionWhere{
				{
					Column:    "name",
					Operation: OpEq,
					Value:     "`John`",
				},
				{
					Operation: "and",
				},
				{
					Column:    "age",
					Operation: OpGt,
					Value:     "30",
				},
			},
		}
		query, args, joins, err := WhereFilter(quoteTo, expr, fieldSearchable)
		assert.NoError(t, err)
		assert.Equal(t, "user_name = ? AND user_age > ?", query)
		assert.Equal(t, []interface{}{"John", 30}, args)
		assert.Equal(t, emptyJoins, joins)
	})

	// Nested expression test
	t.Run("nested expressions", func(t *testing.T) {
		fieldSearchable := FieldSearchable{
			"a": {Column: "a", Type: Int},
			"b": {Column: "b", Type: String},
			"c": {Column: "c", Type: Int},
			"d": {Column: "d", Type: Int},
			"e": {Column: "e", Type: Int},
			"f": {Column: "f", Type: Int},
			"g": {Column: "g", Type: Int},
			"h": {Column: "h", Type: Int},
			"i": {Column: "i", Type: Int},
		}

		expr := &ExpressionWhere{
			Expressions: []ExpressionWhere{
				{Column: "a", Operation: "=", Value: "1"},
				{Operation: "and"},
				{
					Expressions: []ExpressionWhere{
						{Column: "b", Operation: "co", Value: "`test`"},
						{Operation: "or"},
						{Column: "c", Operation: "=", Value: "10"},
					},
				},
				{Operation: "or"},
				{
					Expressions: []ExpressionWhere{
						{Column: "d", Operation: "=", Value: "20"},
						{Operation: "and"},
						{
							Expressions: []ExpressionWhere{
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
		}
		query, args, joins, err := WhereFilter(quoteTo, expr, fieldSearchable)
		assert.NoError(t, err)
		assert.Equal(t, "a = ? AND (b LIKE ? OR c = ?) OR (d = ? AND (e > ? AND f < ? AND (g IS NULL OR g = 0)) OR (h IS NOT NULL AND h <> 0) AND i >= ?)", query)
		assert.Equal(t, []interface{}{1, "%test%", 10, 20, 30, 40, 20}, args)
		assert.Equal(t, emptyJoins, joins)
	})

	// Invalid operator inside nested expressions
	t.Run("invalid operator in nested expressions", func(t *testing.T) {
		expr := &ExpressionWhere{
			Expressions: []ExpressionWhere{
				{Column: "name", Operation: "invalid_op", Value: "`test`"},
			},
		}
		_, _, _, err := WhereFilter(quoteTo, expr, fieldSearchable)
		if !apperr.Is(err, ErrFilterUnknownOperator) {
			t.Errorf("expected %s, actual %s", ErrFilterUnknownOperator, err)
		}
	})

	// Test with bool column
	t.Run("bool filter", func(t *testing.T) {
		expr := &ExpressionWhere{
			Column:    "active",
			Operation: OpNe,
			Value:     "true",
		}
		query, args, joins, err := WhereFilter(quoteTo, expr, fieldSearchable)
		assert.NoError(t, err)
		assert.Equal(t, "is_active = ?", query)
		assert.Equal(t, true, args[0])
		assert.Equal(t, emptyJoins, joins)
	})

	// Test with DateTime filter
	t.Run("datetime filter", func(t *testing.T) {
		expr := &ExpressionWhere{
			Column:    "date",
			Operation: OpGt,
			Value:     "`2022-01-01 00:00:00`",
		}
		query, args, joins, err := WhereFilter(quoteTo, expr, fieldSearchable)
		assert.NoError(t, err)
		assert.Equal(t, "created_at > ?", query)
		assert.Equal(t, "2022-01-01 00:00:00", args[0])
		assert.Equal(t, emptyJoins, joins)
	})

	// Test with empty expression
	t.Run("empty expression", func(t *testing.T) {
		query, args, joins, err := WhereFilter(quoteTo, nil, fieldSearchable)
		assert.NoError(t, err)
		assert.Empty(t, query)
		assert.Empty(t, args)
		assert.Empty(t, joins)
	})

	// Test with empty FieldSearchable
	t.Run("empty FieldSearchable", func(t *testing.T) {
		expr := &ExpressionWhere{Column: "name", Operation: OpEq, Value: "`John`"}
		_, _, _, err := WhereFilter(quoteTo, expr, FieldSearchable{})
		assert.Error(t, err)
	})

	// Test with NP
	t.Run("NP", func(t *testing.T) {
		expr := &ExpressionWhere{Column: "name", Operation: OpNp, Value: ""}
		query, args, joins, err := WhereFilter(quoteTo, expr, fieldSearchable)
		assert.NoError(t, err)
		assert.Equal(t, "(user_name IS NOT NULL AND user_name <> '')", query)
		assert.Equal(t, emptyArgs, args)
		assert.Equal(t, emptyJoins, joins)
	})

	// Test with OpPt
	t.Run("OpPt", func(t *testing.T) {
		expr := &ExpressionWhere{Column: "name", Operation: OpPt, Value: ""}
		query, args, joins, err := WhereFilter(quoteTo, expr, fieldSearchable)
		assert.NoError(t, err)
		assert.Equal(t, "(user_name IS NULL OR user_name = '')", query)
		assert.Equal(t, emptyArgs, args)
		assert.Equal(t, emptyJoins, joins)
	})

	// Test with unknown field type
	t.Run("unknown field type", func(t *testing.T) {
		expr := &ExpressionWhere{Column: "unknown", Operation: OpEq, Value: "`test`"}
		_, _, _, err := WhereFilter(quoteTo, expr, fieldSearchable)
		assert.Error(t, err)
	})

	// Test with bool column
	t.Run("invalid operator", func(t *testing.T) {
		exprs := &ExpressionWhere{
			Expressions: []ExpressionWhere{
				{Column: "active", Operation: OpEq, Value: "`John`"},
				{Operation: "and"},
				{Column: "age", Operation: OpGt, Value: "25"},
			},
		}
		_, _, _, err := WhereFilter(quoteTo, exprs, fieldSearchable)
		if !apperr.Is(err, ErrFilterInvalidValue) {
			t.Errorf("expected %s, actual %s", ErrFilterInvalidValue, err)
		}
	})
}

func TestGenerateWhereClause(t *testing.T) {
	fieldSearchable := FieldSearchable{
		"name": {Column: "user_name", Type: String},
		"age":  {Column: "user_age", Type: Int},
		"id":   {Column: "", Type: Int, SQL: "id = ?"},
	}

	// Test for complex nested expressions
	t.Run("nested expressions with AND", func(t *testing.T) {
		exprs := []ExpressionWhere{
			{Column: "name", Operation: OpEq, Value: "`John`"},
			{Operation: "and"},
			{Column: "age", Operation: OpGt, Value: "25"},
			{Operation: "or"},
			{Column: "id", Operation: OpIn, Value: "43"},
		}
		query, args, joins, err := generateWhereClause(quoteTo, exprs, fieldSearchable)
		assert.NoError(t, err)
		assert.Equal(t, "user_name = ? AND user_age > ? OR id = ?", query)
		assert.Equal(t, []interface{}{"John", 25, 43}, args)
		assert.Empty(t, joins)
	})
}

func TestFormatWhereString(t *testing.T) {
	fieldSearchable := FieldSearchable{
		"name":   {Column: "user_name", Type: String},
		"age":    {Column: "user_age", Type: Int},
		"active": {Column: "is_active", Type: Bool},
		"date":   {Column: "created_at", Type: DateTime},
		"status": {Column: "status", Type: Bool},
	}

	// Test for invalid operator
	t.Run("invalid operator", func(t *testing.T) {
		expr := ExpressionWhere{Column: "name", Operation: "invalid_op", Value: "`John`"}
		_, _, _, err := formatWhereString(quoteTo, expr, fieldSearchable["name"])
		if !apperr.Is(err, ErrFilterUnknownOperator) {
			t.Errorf("expected %s, actual %s", ErrFilterUnknownOperator, err)
		}
	})

	t.Run(OpEq, func(t *testing.T) {
		expr := ExpressionWhere{Column: "name", Operation: OpEq, Value: "`John`"}
		query, args, joins, err := formatWhereString(quoteTo, expr, fieldSearchable["name"])
		assert.NoError(t, err)
		assert.Equal(t, "user_name = ?", query)
		assert.Equal(t, "John", args)
		assert.Empty(t, joins)
	})
	t.Run(OpCo, func(t *testing.T) {
		expr := ExpressionWhere{Column: "name", Operation: OpCo, Value: "`John`"}
		query, args, joins, err := formatWhereString(quoteTo, expr, fieldSearchable["name"])
		assert.NoError(t, err)
		assert.Equal(t, "user_name LIKE ?", query)
		assert.Equal(t, "%John%", args)
		assert.Empty(t, joins)
	})
	t.Run(OpSw, func(t *testing.T) {
		expr := ExpressionWhere{Column: "name", Operation: OpSw, Value: "`John`"}
		query, args, joins, err := formatWhereString(quoteTo, expr, fieldSearchable["name"])
		assert.NoError(t, err)
		assert.Equal(t, "user_name LIKE ?", query)
		assert.Equal(t, "John%", args)
		assert.Empty(t, joins)
	})
	t.Run(OpEw, func(t *testing.T) {
		expr := ExpressionWhere{Column: "name", Operation: OpEw, Value: "`John`"}
		query, args, joins, err := formatWhereString(quoteTo, expr, fieldSearchable["name"])
		assert.NoError(t, err)
		assert.Equal(t, "user_name LIKE ?", query)
		assert.Equal(t, "%John", args)
		assert.Empty(t, joins)
	})
	t.Run(OpPt, func(t *testing.T) {
		expr := ExpressionWhere{Column: "name", Operation: OpPt, Value: "`John`"}
		query, args, joins, err := formatWhereString(quoteTo, expr, fieldSearchable["name"])
		assert.NoError(t, err)
		assert.Equal(t, "(user_name IS NULL OR user_name = '')", query)
		assert.Empty(t, args)
		assert.Empty(t, joins)
	})
	t.Run(OpNp, func(t *testing.T) {
		expr := ExpressionWhere{Column: "name", Operation: OpNp, Value: "`John`"}
		query, args, joins, err := formatWhereString(quoteTo, expr, fieldSearchable["name"])
		assert.NoError(t, err)
		assert.Equal(t, "(user_name IS NOT NULL AND user_name <> '')", query)
		assert.Empty(t, args)
		assert.Empty(t, joins)
	})
	t.Run(OpIn, func(t *testing.T) {
		expr := ExpressionWhere{Column: "name", Operation: OpIn, Value: "`John`,`John2`"}
		query, args, joins, err := formatWhereString(quoteTo, expr, fieldSearchable["name"])
		assert.NoError(t, err)
		assert.Equal(t, "user_name IN ?", query)
		assert.Equal(t, []string{"John", "John2"}, args)
		assert.Empty(t, joins)
	})
	t.Run(OpIn, func(t *testing.T) {
		expr := ExpressionWhere{Column: "status", Operation: OpIn, Value: "true,false"}
		query, args, joins, err := formatWhereString(quoteTo, expr, fieldSearchable["status"])
		assert.NoError(t, err)
		assert.Equal(t, "status IN ?", query)
		assert.Equal(t, []bool{true, false}, args)
		assert.Empty(t, joins)
	})
	t.Run(OpNin, func(t *testing.T) {
		expr := ExpressionWhere{Column: "status", Operation: OpIn, Value: "true"}
		query, args, joins, err := formatWhereString(quoteTo, expr, fieldSearchable["status"])
		assert.NoError(t, err)
		assert.Equal(t, "status NOT IN ?", query)
		assert.Equal(t, []bool{true}, args)
		assert.Empty(t, joins)
	})
	t.Run(OpEq, func(t *testing.T) {
		expr := ExpressionWhere{Column: "status", Operation: OpEq, Value: "true"}
		query, args, joins, err := formatWhereString(quoteTo, expr, fieldSearchable["status"])
		assert.NoError(t, err)
		assert.Equal(t, "status = ?", query)
		assert.Equal(t, true, args)
		assert.Empty(t, joins)
	})
}

func TestFormatIntWhere(t *testing.T) {
	t.Run("valid int filter", func(t *testing.T) {
		expr := ExpressionWhere{Column: "age", Operation: OpEq, Value: "30"}
		query, arg, err := formatIntWhere(expr, "user_age")
		assert.NoError(t, err)
		assert.Equal(t, "user_age = ?", query)
		assert.Equal(t, 30, arg)
	})

	t.Run("invalid int value", func(t *testing.T) {
		expr := ExpressionWhere{Column: "age", Operation: OpEq, Value: "invalid"}
		_, _, err := formatIntWhere(expr, "`user_age`")
		if !apperr.Is(err, ErrFilterInvalidValue) {
			t.Errorf("expected %s, actual %s", ErrFilterInvalidValue, err)
		}
	})

	t.Run(OpGt, func(t *testing.T) {
		expr := ExpressionWhere{Column: "age", Operation: OpGt, Value: "30"}
		query, arg, err := formatIntWhere(expr, "user_age")
		assert.NoError(t, err)
		assert.Equal(t, "user_age > ?", query)
		assert.Equal(t, 30, arg)
	})

	t.Run(OpLt, func(t *testing.T) {
		expr := ExpressionWhere{Column: "age", Operation: OpLt, Value: "30"}
		query, arg, err := formatIntWhere(expr, "user_age")
		assert.NoError(t, err)
		assert.Equal(t, "user_age < ?", query)
		assert.Equal(t, 30, arg)
	})

	t.Run(OpGte, func(t *testing.T) {
		expr := ExpressionWhere{Column: "age", Operation: OpGte, Value: "30"}
		query, arg, err := formatIntWhere(expr, "user_age")
		assert.NoError(t, err)
		assert.Equal(t, "user_age >= ?", query)
		assert.Equal(t, 30, arg)
	})

	t.Run(OpLte, func(t *testing.T) {
		expr := ExpressionWhere{Column: "age", Operation: OpLte, Value: "30"}
		query, arg, err := formatIntWhere(expr, "user_age")
		assert.NoError(t, err)
		assert.Equal(t, "user_age <= ?", query)
		assert.Equal(t, 30, arg)
	})

	t.Run(OpEq, func(t *testing.T) {
		expr := ExpressionWhere{Column: "age", Operation: OpEq, Value: "30"}
		query, arg, err := formatIntWhere(expr, "user_age")
		assert.NoError(t, err)
		assert.Equal(t, "user_age = ?", query)
		assert.Equal(t, 30, arg)
	})

	t.Run(OpPt, func(t *testing.T) {
		expr := ExpressionWhere{Column: "age", Operation: OpPt, Value: "30"}
		query, arg, err := formatIntWhere(expr, "user_age")
		assert.NoError(t, err)
		assert.Equal(t, "(user_age IS NULL OR user_age = 0)", query)
		assert.Empty(t, arg)
	})

	t.Run(OpNp, func(t *testing.T) {
		expr := ExpressionWhere{Column: "age", Operation: OpNp, Value: "30"}
		query, arg, err := formatIntWhere(expr, "user_age")
		assert.NoError(t, err)
		assert.Equal(t, "(user_age IS NOT NULL AND user_age <> 0)", query)
		assert.Empty(t, arg)
	})

	t.Run(OpIn, func(t *testing.T) {
		expr := ExpressionWhere{Column: "age", Operation: OpIn, Value: "30,40"}
		query, arg, err := formatIntWhere(expr, "user_age")
		assert.NoError(t, err)
		assert.Equal(t, "user_age IN ?", query)
		assert.Equal(t, []int{30, 40}, arg)
	})
}

func TestFormatBoolWhere(t *testing.T) {
	t.Run("valid bool filter", func(t *testing.T) {
		expr := ExpressionWhere{Column: "active", Operation: OpNe, Value: "true"}
		query, arg, err := formatBoolWhere(expr, "is_active")
		assert.NoError(t, err)
		assert.Equal(t, "is_active = ?", query)
		assert.Equal(t, true, arg)
	})

	t.Run("invalid bool value", func(t *testing.T) {
		expr := ExpressionWhere{Column: "active", Operation: OpNe, Value: "invalid"}
		_, _, err := formatBoolWhere(expr, "is_active")
		if !apperr.Is(err, ErrFilterInvalidValue) {
			t.Errorf("expected %s, actual %s", ErrFilterInvalidValue, err)
		}
	})
}

func TestFormatDateTimeWhere(t *testing.T) {
	t.Run("valid datetime filter", func(t *testing.T) {
		expr := ExpressionWhere{Column: "date", Operation: OpGt, Value: "`2022-01-01 00:00:00`"}
		query, arg, err := formatDateTimeWhere(expr, "created_at")
		assert.NoError(t, err)
		assert.Equal(t, "created_at > ?", query)
		assert.Equal(t, "2022-01-01 00:00:00", arg)
	})

	t.Run("invalid datetime format", func(t *testing.T) {
		expr := ExpressionWhere{Column: "date", Operation: OpGt, Value: "`invalid_date`"}
		_, _, err := formatDateTimeWhere(expr, "created_at")
		if !apperr.Is(err, ErrFilterInvalidValue) {
			t.Errorf("expected %s, actual %s", ErrFilterInvalidValue, err)
		}
	})
}

func TestFormatSQLWhere(t *testing.T) {
	t.Run("valid sql filter", func(t *testing.T) {
		expr := ExpressionWhere{Column: "date", Operation: OpIn, Value: "1,2"}
		query, arg, err := formatSQLWhere(expr, "a = ? AND b = ?")
		assert.NoError(t, err)
		assert.Equal(t, "a = ? AND b = ?", query)
		assert.Equal(t, []interface{}{1, 2}, arg)
	})

	t.Run("invalid sql operation", func(t *testing.T) {
		expr := ExpressionWhere{Column: "date", Operation: OpPt, Value: "1"}
		_, _, err := formatSQLWhere(expr, "a = ? AND b = ?")
		if !apperr.Is(err, ErrFilterUnknownOperator) {
			t.Errorf("expected %s, actual %s", ErrFilterUnknownOperator, err)
		}
	})

	t.Run("invalid sql count <", func(t *testing.T) {
		expr := ExpressionWhere{Column: "date", Operation: OpIn, Value: "1"}
		_, _, err := formatSQLWhere(expr, "a = ? AND b = ?")
		if !apperr.Is(err, ErrFilterInvalidValue) {
			t.Errorf("expected %s, actual %s", ErrFilterInvalidValue, err)
		}
	})

	t.Run("invalid sql count >", func(t *testing.T) {
		expr := ExpressionWhere{Column: "date", Operation: OpIn, Value: "1,2,3"}
		_, _, err := formatSQLWhere(expr, "a = ? AND b = ?")
		if !apperr.Is(err, ErrFilterInvalidValue) {
			t.Errorf("expected %s, actual %s", ErrFilterInvalidValue, err)
		}
	})

	t.Run("invalid sql format", func(t *testing.T) {
		expr := ExpressionWhere{Column: "date", Operation: OpIn, Value: "invalid"}
		_, _, err := formatSQLWhere(expr, "a = ?")
		if !apperr.Is(err, ErrFilterInvalidValue) {
			t.Errorf("expected %s, actual %s", ErrFilterInvalidValue, err)
		}
	})
}
