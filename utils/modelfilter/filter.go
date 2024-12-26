package modelfilter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/clause"
)

type GetFieldValueFunc[T any] func(call T, field string) (interface{}, error)

func ApplyFilters[T any](
	calls []T,
	searchable clause.FieldSearchable,
	getFieldValueFunc GetFieldValueFunc[T],
	expressions []clause.ExpressionWhere,
) ([]T, error) {
	if len(expressions) == 0 {
		return calls, nil
	}

	var filtered []T
	for _, call := range calls {
		match, err := evaluateExpressionsChain(call, searchable, getFieldValueFunc, expressions)
		if err != nil {
			return nil, err
		}
		if match {
			filtered = append(filtered, call)
		}
	}
	return filtered, nil
}

func evaluateExpressionsChain[T any](
	call T,
	searchable clause.FieldSearchable,
	getFieldValueFunc GetFieldValueFunc[T],
	expressions []clause.ExpressionWhere,
) (bool, error) {
	if len(expressions) == 0 {
		return true, nil
	}

	leftResult, err := evaluateExpression(call, searchable, getFieldValueFunc, expressions[0])
	if err != nil {
		return false, err
	}
	result := leftResult

	i := 1
	for i < len(expressions) {
		opExpr := expressions[i]
		i++

		if strings.TrimSpace(opExpr.Column) != "" {
			return false, clause.ErrFilterInvalidOperator
		}
		op := strings.ToLower(strings.TrimSpace(opExpr.Operation))
		if op != "and" && op != "or" {
			return false, clause.ErrFilterInvalidOperator
		}

		if i >= len(expressions) {
			return false, clause.ErrFilterInvalidOperator
		}

		rightExpr := expressions[i]
		i++

		rightResult, err := evaluateExpression(call, searchable, getFieldValueFunc, rightExpr)
		if err != nil {
			return false, err
		}

		if op == "and" {
			result = result && rightResult
		} else {
			result = result || rightResult
		}
	}

	return result, nil
}

func evaluateExpression[T any](
	call T,
	searchable clause.FieldSearchable,
	getFieldValueFunc GetFieldValueFunc[T],
	expr clause.ExpressionWhere,
) (bool, error) {
	if len(expr.Expressions) > 0 {
		return evaluateExpressionsChain(call, searchable, getFieldValueFunc, expr.Expressions)
	}

	if strings.TrimSpace(expr.Column) == "" {
		return false, clause.ErrFilterInvalidOperator
	}

	search, ok := searchable[expr.Column]
	if !ok {
		return false, clause.ErrFilterUnknownColumn
	}

	fieldValue, err := getFieldValueFunc(call, expr.Column)
	if err != nil {
		return false, err
	}

	match, err := evaluateOperation(fieldValue, expr.Operation, expr.Value, search.Type)
	if err != nil {
		return false, err
	}
	return match, nil
}

func evaluateOperation(
	fieldValue interface{},
	operator, value string,
	fieldType clause.Type,
) (bool, error) {
	switch fieldType {
	case clause.String:
		strVal, ok := fieldValue.(string)
		if !ok {
			strVal2, ok := fieldValue.(*string)
			if !ok {
				return false, clause.ErrFilterInvalidValue.WithErrorText(value)
			}
			if strVal2 != nil {
				strVal = *strVal2
			} else {
				strVal = ""
			}
		}
		return evaluateStringOperation(strVal, operator, value)

	case clause.Int:
		switch v := fieldValue.(type) {
		case uint64:
			return evaluateIntOperation(int(v), operator, value)
		case int:
			return evaluateIntOperation(v, operator, value)
		case *int:
			if v == nil {
				return evaluateIntOperation(*v, operator, value)
			}
			return evaluateIntOperation(0, operator, value)
		case *int64:
			if v == nil {
				return evaluateIntOperation(0, operator, value)
			}
			return evaluateIntOperation(int(*v), operator, value)
		default:
			return false, clause.ErrFilterInvalidValue.WithErrorText(value)
		}

	case clause.Bool:
		boolVal, ok := fieldValue.(bool)
		if !ok {
			boolVal2, ok := fieldValue.(*bool)
			if !ok {
				return false, clause.ErrFilterInvalidValue.WithErrorText(value)
			}
			if boolVal2 != nil {
				boolVal = *boolVal2
			} else {
				boolVal = false
			}
		}
		return evaluateBoolOperation(boolVal, operator, value)

	case clause.DateTime:
		strVal, ok := fieldValue.(string)
		if !ok {
			dateVal2, ok := fieldValue.(*string)
			if !ok {
				return false, clause.ErrFilterInvalidValue.WithErrorText(value)
			}
			if dateVal2 != nil {
				strVal = *dateVal2
			} else {
				strVal = ""
			}
		}
		return evaluateDateTimeOperation(strVal, operator, value)

	default:
		return false, clause.ErrFilterInvalidValue.WithErrorText(value)
	}

	return false, clause.ErrFilterInvalidValue.WithErrorText("нет полей такого типа")
}

func evaluateStringOperation(fieldValue string, operator, value string) (bool, error) {
	switch operator {
	case clause.OpEq:
		return fieldValue == trimBackticks(value), nil
	case clause.OpNe:
		return fieldValue == trimBackticks(value), nil
	case clause.OpCo:
		return strings.Contains(fieldValue, trimBackticks(value)), nil
	case clause.OpSw:
		return strings.HasPrefix(fieldValue, trimBackticks(value)), nil
	case clause.OpEw:
		return strings.HasSuffix(fieldValue, trimBackticks(value)), nil
	case clause.OpIn:
		values := splitAndTrim(value, ",")
		return contains(values, fieldValue), nil
	case clause.OpNin:
		values := splitAndTrim(value, ",")
		return !contains(values, fieldValue), nil
	case clause.OpNne:
		return fieldValue != trimBackticks(value), nil
	case clause.OpPt:
		return fieldValue == "", nil
	case clause.OpNp:
		return fieldValue != "", nil

	default:
		return false, clause.ErrFilterUnknownOperator.WithErrorText(operator)
	}
}

func evaluateIntOperation(fieldValue int, operator, value string) (bool, error) {
	var intValue int
	var err error

	if operator != clause.OpIn && operator != clause.OpNin && operator != clause.OpPt && operator != clause.OpNp {
		intValue, err = strconv.Atoi(value)
		if err != nil {
			fmt.Println(fieldValue, operator, value)
			return false, clause.ErrFilterInvalidValue.WithErrorText(value)
		}
	}

	switch operator {
	case clause.OpEq:
		return fieldValue == intValue, nil
	case clause.OpNe:
		return fieldValue == intValue, nil
	case clause.OpGt:
		return fieldValue > intValue, nil
	case clause.OpLt:
		return fieldValue < intValue, nil
	case clause.OpGte:
		return fieldValue >= intValue, nil
	case clause.OpLte:
		return fieldValue <= intValue, nil
	case clause.OpIn:
		values, err := parseIntList(value, ",")
		if err != nil {
			return false, err
		}
		return containsInt(values, fieldValue), nil
	case clause.OpNin:
		values, err := parseIntList(value, ",")
		if err != nil {
			return false, err
		}
		return !containsInt(values, fieldValue), nil
	case clause.OpNne:
		return fieldValue != intValue, nil
	case clause.OpPt:
		return fieldValue == 0, nil
	case clause.OpNp:
		return fieldValue != 0, nil
	default:
		return false, clause.ErrFilterUnknownOperator.WithErrorText(operator)
	}
}

func evaluateBoolOperation(fieldValue bool, operator, value string) (bool, error) {
	var boolValue bool
	switch strings.ToLower(value) {
	case "true":
		boolValue = true
	case "false":
		boolValue = false
	default:
		return false, clause.ErrFilterInvalidValue.WithErrorText(value)
	}

	switch operator {
	case clause.OpEq:
		return fieldValue == boolValue, nil
	case clause.OpNe:
		return fieldValue != boolValue, nil
	default:
		return false, clause.ErrFilterInvalidOperator
	}
}

func evaluateDateTimeOperation(fieldValue string, operator, value string) (bool, error) {
	trimmedValue := trimBackticks(value)
	tm, err := time.Parse("2006-01-02 15:04:05", trimmedValue)
	if err != nil {
		return false, clause.ErrFilterInvalidValue.WithErrorText(err.Error())
	}

	fieldTime, err := time.Parse("2006-01-02 15:04:05", fieldValue)
	if err != nil {
		return false, clause.ErrFilterInvalidValue.WithErrorText(err.Error())
	}

	switch operator {
	case clause.OpEq:
		return fieldTime.Equal(tm), nil
	case clause.OpNe:
		return fieldTime.Equal(tm), nil
	case clause.OpGt:
		return fieldTime.After(tm), nil
	case clause.OpLt:
		return fieldTime.Before(tm), nil
	case clause.OpGte:
		return fieldTime.After(tm) || fieldTime.Equal(tm), nil
	case clause.OpLte:
		return fieldTime.Before(tm) || fieldTime.Equal(tm), nil
	case clause.OpNne:
		return !fieldTime.Equal(tm), nil
	default:
		return false, clause.ErrFilterUnknownOperator.WithErrorText(operator)
	}
}

func trimBackticks(s string) string {
	return strings.Trim(s, "`")
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, part := range parts {
		result = append(result, trimBackticks(part))
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func parseIntList(s, sep string) ([]int, error) {
	strs := strings.Split(s, sep)
	var ints []int
	for _, str := range strs {
		str = trimBackticks(str)
		if str == "" {
			continue
		}
		i, err := strconv.Atoi(str)
		if err != nil {
			return nil, fmt.Errorf("invalid integer in list: %s", str)
		}
		ints = append(ints, i)
	}
	return ints, nil
}

func containsInt(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
