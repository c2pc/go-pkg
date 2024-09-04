package clause

import (
	"encoding/json"
	"fmt"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/i18n"
	"strconv"
	"strings"
	"time"
)

var (
	ErrFilterUnknownOperator = apperr.New("filter_unknown_operator", apperr.WithTextTranslate(i18n.ErrFilterUnknownOperator), apperr.WithCode(code.InvalidArgument))
	ErrFilterInvalidOperator = apperr.New("filter_invalid_operator", apperr.WithTextTranslate(i18n.ErrFilterInvalidOperator), apperr.WithCode(code.InvalidArgument))
	ErrFilterUnknownColumn   = apperr.New("filter_unknown_column", apperr.WithTextTranslate(i18n.ErrFilterUnknownColumn), apperr.WithCode(code.InvalidArgument))
	ErrFilterInvalidValue    = apperr.New("filter_invalid_value", apperr.WithTextTranslate(i18n.ErrFilterInvalidValue), apperr.WithCode(code.InvalidArgument))
)

type ExpressionWhere struct {
	Expressions []ExpressionWhere
	Column      string
	Operation   string
	Value       string
}

type Type string

const (
	String   Type = "string"
	Int      Type = "int"
	Bool     Type = "bool"
	DateTime Type = "datetime"
)

type FieldSearchable map[string]Search

type Search struct {
	Column string
	Type   Type
	Join   string
}

func WhereFilter(quoteTo func(string) string, expression *ExpressionWhere, fieldSearchable FieldSearchable) (string, []interface{}, []string, error) {
	if expression == nil {
		return "", nil, nil, nil
	}

	if len(expression.Expressions) == 0 {
		expression.Expressions = []ExpressionWhere{
			{Column: expression.Column, Operation: expression.Operation, Value: expression.Value},
		}
	}

	b, err := json.Marshal(expression.Expressions)
	fmt.Println(string(b))

	query, args, joins, err := generateWhereClause(quoteTo, expression.Expressions, fieldSearchable)
	if err != nil {
		return "", nil, nil, err
	}

	return query, args, joins, nil
}

func generateWhereClause(quoteTo func(string) string, exprs []ExpressionWhere, fieldSearchable FieldSearchable) (string, []interface{}, []string, error) {
	var result []string
	var joins []string
	var values []interface{}

	for i, expr := range exprs {
		if expr.Column != "" {
			query, args, join, err := formatWhereString(quoteTo, expr, fieldSearchable)
			if err != nil {
				return "", nil, nil, err
			}
			result = append(result, query)
			if args != nil {
				values = append(values, args)
			}
			if join != "" {
				joins = append(joins, join)
			}
		} else if expr.Operation != "" {
			// Добавляем логические операторы
			if expr.Operation == "and" || expr.Operation == "or" {
				if i == len(exprs)-1 {
					return "", nil, nil, ErrFilterInvalidOperator.WithTextArgs(expr.Operation)
				}
				result = append(result, fmt.Sprintf(" %s ", strings.ToUpper(expr.Operation)))
			}
		}
		if len(expr.Expressions) > 0 {
			// Рекурсивный вызов для вложенных выражений
			subQuery, subValues, subJoins, err := generateWhereClause(quoteTo, expr.Expressions, fieldSearchable)
			if err != nil {
				return "", nil, nil, err
			}
			result = append(result, "("+subQuery+")")
			values = append(values, subValues...)
			joins = append(joins, subJoins...)
		}
	}
	return strings.Join(result, ""), values, joins, nil
}

func formatWhereString(quoteTo func(string) string, expr ExpressionWhere, fieldSearchable FieldSearchable) (string, interface{}, string, error) {
	if search, ok := fieldSearchable[expr.Column]; ok {
		column := upperModels(quoteTo(search.Column))
		join := quoteTo(search.Join)
		switch search.Type {
		case String:
			var values []string
			if expr.Operation != "in" && expr.Operation != "pt" && expr.Operation != "np" {
				if len(expr.Value) < 3 || expr.Value[0] != '`' || expr.Value[len(expr.Value)-1] != '`' {
					return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column)
				}
				val := strings.ReplaceAll(expr.Value, "`", "")
				values = []string{val}
			} else if expr.Operation == "in" {
				values = []string{}
				vals := strings.Split(expr.Value, "`,`")
				for _, val := range vals {
					val = strings.ReplaceAll(val, "`", "")
					values = append(values, val)
				}
				if len(values) == 0 {
					return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column)
				}
			}

			switch expr.Operation {
			case "co":
				return fmt.Sprintf("%s LIKE ?", column), "%" + values[0] + "%", join, nil
			case "eq":
				return fmt.Sprintf("%s = ?", column), values[0], join, nil
			case "sw":
				return fmt.Sprintf("%s LIKE ?", column), "%" + values[0], join, nil
			case "ew":
				return fmt.Sprintf("%s LIKE ?", column), values[0] + "%", join, nil
			case "pt":
				return fmt.Sprintf("(%s IS NULL OR %s = '')", column, column), nil, join, nil
			case "np":
				return fmt.Sprintf("(%s IS NOT NULL AND %s <> '')", column, column), nil, join, nil
			case "in":
				return fmt.Sprintf("%s IN ?", column), values, join, nil
			default:
				return "", nil, "", ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
			}
		case Int:
			var values []int
			if expr.Operation != "in" && expr.Operation != "pt" && expr.Operation != "np" {
				val, err := strconv.Atoi(expr.Value)
				if err != nil {
					return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column)
				}
				values = []int{val}
			} else if expr.Operation == "in" {
				values = []int{}
				vals := strings.Split(expr.Value, ",")
				for _, val := range vals {
					v, err := strconv.Atoi(val)
					if err != nil {
						return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithError(err)
					}
					values = append(values, v)
				}
				if len(values) == 0 {
					return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column)
				}
			}
			switch expr.Operation {
			case ">":
				return fmt.Sprintf("%s > ?", column), values[0], join, nil
			case "<":
				return fmt.Sprintf("%s < ?", column), values[0], join, nil
			case ">=":
				return fmt.Sprintf("%s >= ?", column), values[0], join, nil
			case "<=":
				return fmt.Sprintf("%s <= ?", column), values[0], join, nil
			case "=":
				return fmt.Sprintf("%s = ?", column), values[0], join, nil
			case "pt":
				return fmt.Sprintf("(%s IS NULL OR %s = 0)", column, column), nil, join, nil
			case "np":
				return fmt.Sprintf("(%s IS NOT NULL AND %s <> 0)", column, column), nil, join, nil
			case "in":
				return fmt.Sprintf("%s IN ?", column), values, join, nil
			default:
				return "", nil, "", ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
			}
		case Bool:
			var value bool
			if expr.Value == "true" {
				value = true
			} else if expr.Value == "false" {
				value = false
			} else {
				return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column)
			}
			switch expr.Operation {
			case "=":
				return fmt.Sprintf("%s = ?", column), value, join, nil
			default:
				return "", nil, "", ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
			}

		case DateTime:
			var value string
			if len(expr.Value) < 3 || expr.Value[0] != '`' || expr.Value[len(expr.Value)-1] != '`' {
				return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column)
			}
			tm, err := time.Parse("2006-01-02 15:04:05", strings.ReplaceAll(expr.Value, "`", ""))
			if err != nil {
				return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithError(err)
			}

			value = tm.Format("2006-01-02 15:04:05")

			switch expr.Operation {
			case ">":
				return fmt.Sprintf("%s > ?", column), value, join, nil
			case "<":
				return fmt.Sprintf("%s < ?", column), value, join, nil
			case ">=":
				return fmt.Sprintf("%s >= ?", column), value, join, nil
			case "<=":
				return fmt.Sprintf("%s <= ?", column), value, join, nil
			case "=":
				return fmt.Sprintf("%s = ?", column), value, join, nil
			default:
				return "", nil, "", ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
			}
		default:
			return "", nil, "", fmt.Errorf("unknown type %s for %s", search.Type, expr.Column)
		}
	}

	return "", nil, "", ErrFilterUnknownColumn.WithTextArgs(expr.Column)
}
