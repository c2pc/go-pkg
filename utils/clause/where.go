package clause

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/i18n"
)

// Ошибки для фильтрации
var (
	ErrFilterUnknownOperator = apperr.New("filter_unknown_operator", apperr.WithTextTranslate(i18n.ErrFilterUnknownOperator), apperr.WithCode(code.InvalidArgument))
	ErrFilterInvalidOperator = apperr.New("filter_invalid_operator", apperr.WithTextTranslate(i18n.ErrFilterInvalidOperator), apperr.WithCode(code.InvalidArgument))
	ErrFilterUnknownColumn   = apperr.New("filter_unknown_column", apperr.WithTextTranslate(i18n.ErrFilterUnknownColumn), apperr.WithCode(code.InvalidArgument))
	ErrFilterInvalidValue    = apperr.New("filter_invalid_value", apperr.WithTextTranslate(i18n.ErrFilterInvalidValue), apperr.WithCode(code.InvalidArgument))
)

// ExpressionWhere описывает одно выражение для фильтрации
type ExpressionWhere struct {
	Expressions []ExpressionWhere `json:"expressions"` // Вложенные выражения для составных условий
	Column      string            `json:"column"`      // Имя столбца для фильтрации
	Operation   string            `json:"operation"`   // Операция фильтрации
	Value       string            `json:"value"`       // Значение для фильтрации
}

// Type представляет тип данных для столбца
type Type string

const (
	String   Type = "string"   // Строка
	Int      Type = "int"      // Целое число
	Bool     Type = "bool"     // Логическое значение
	DateTime Type = "datetime" // Дата и время
)

// Константы операций фильтрации
const (
	OpIn  = "in" // Операция "IN"
	OpPt  = "pt" // Операция "IS NULL OR = ''"
	OpNp  = "np" // Операция "IS NOT NULL AND <> ''"
	OpCo  = "co" // Операция "LIKE %...%"
	OpEq  = "eq" // Операция "="
	OpSw  = "sw" // Операция "LIKE ...%"
	OpEw  = "ew" // Операция "LIKE %..."
	OpGt  = ">"  // Операция ">"
	OpLt  = "<"  // Операция "<"
	OpGte = ">=" // Операция ">="
	OpLte = "<=" // Операция "<="
	OpNe  = "="  // Операция "="
)

// FieldSearchable представляет карту столбцов с информацией о том, как они можно искать
type FieldSearchable map[string]Search

// Search содержит информацию о том, как искать в определенном столбце
type Search struct {
	Column string // Имя столбца в базе данных
	Type   Type   // Тип данных в столбце
	Join   string // Имя соединения для запроса
}

// WhereFilter строит SQL-запрос на основе выражения фильтрации
func WhereFilter(quoteTo func(string) string, expression *ExpressionWhere, fieldSearchable FieldSearchable) (string, []interface{}, []string, error) {
	if expression == nil {
		return "", nil, nil, nil
	}

	if len(expression.Expressions) == 0 {
		// Если выражение не содержит вложенных выражений, создаем одно базовое
		expression.Expressions = []ExpressionWhere{
			{Column: expression.Column, Operation: expression.Operation, Value: expression.Value},
		}
	}

	// Генерируем SQL-запрос и параметры
	query, args, joins, err := generateWhereClause(quoteTo, expression.Expressions, fieldSearchable)
	if err != nil {
		return "", nil, nil, err
	}

	return query, args, joins, nil
}

// generateWhereClause строит SQL-запрос из массива выражений
func generateWhereClause(quoteTo func(string) string, exprs []ExpressionWhere, fieldSearchable FieldSearchable) (string, []interface{}, []string, error) {
	var result []string
	var joins []string
	var values []interface{}

	for i, expr := range exprs {
		if expr.Column != "" {
			// Форматируем выражение для столбца
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
			// Добавляем логические операторы "AND" или "OR"
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

// formatWhereString форматирует строку для SQL-запроса в зависимости от типа данных
func formatWhereString(quoteTo func(string) string, expr ExpressionWhere, fieldSearchable FieldSearchable) (string, interface{}, string, error) {
	if search, ok := fieldSearchable[expr.Column]; ok {
		column := upperModels(quoteTo(search.Column))
		join := quoteTo(search.Join)
		switch search.Type {
		case String:
			return formatStringWhere(expr, column, join)
		case Int:
			return formatIntWhere(expr, column, join)
		case Bool:
			return formatBoolWhere(expr, column, join)
		case DateTime:
			return formatDateTimeWhere(expr, column, join)
		default:
			return "", nil, "", fmt.Errorf("unknown type %s for %s", search.Type, expr.Column)
		}
	}

	return "", nil, "", ErrFilterUnknownColumn.WithTextArgs(expr.Column)
}

// formatStringWhere форматирует условие для строковых столбцов
func formatStringWhere(expr ExpressionWhere, column string, join string) (string, interface{}, string, error) {
	var values []string
	if expr.Operation != OpIn && expr.Operation != OpPt && expr.Operation != OpNp {
		if len(expr.Value) < 3 || expr.Value[0] != '`' || expr.Value[len(expr.Value)-1] != '`' {
			return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column)
		}
		val := strings.ReplaceAll(expr.Value, "`", "")
		values = []string{val}
	} else if expr.Operation == OpIn {
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
	case OpEq:
		return fmt.Sprintf("%s = ?", column), values[0], join, nil
	case OpCo:
		return fmt.Sprintf("%s LIKE ?", column), "%" + values[0] + "%", join, nil
	case OpSw:
		return fmt.Sprintf("%s LIKE ?", column), values[0] + "%", join, nil
	case OpEw:
		return fmt.Sprintf("%s LIKE ?", column), "%" + values[0], join, nil
	case OpPt:
		return fmt.Sprintf("(%s IS NULL OR %s = '')", column, column), nil, join, nil
	case OpNp:
		return fmt.Sprintf("(%s IS NOT NULL AND %s <> '')", column, column), nil, join, nil
	case OpIn:
		return fmt.Sprintf("%s IN ?", column), values, join, nil
	default:
		return "", nil, "", ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
	}
}

// formatIntWhere форматирует условие для целочисленных столбцов
func formatIntWhere(expr ExpressionWhere, column string, join string) (string, interface{}, string, error) {
	var values []int
	if expr.Operation != OpIn && expr.Operation != OpPt && expr.Operation != OpNp {
		val, err := strconv.Atoi(expr.Value)
		if err != nil {
			return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithError(err)
		}
		values = []int{val}
	} else if expr.Operation == OpIn {
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
	case OpGt:
		return fmt.Sprintf("%s > ?", column), values[0], join, nil
	case OpLt:
		return fmt.Sprintf("%s < ?", column), values[0], join, nil
	case OpGte:
		return fmt.Sprintf("%s >= ?", column), values[0], join, nil
	case OpLte:
		return fmt.Sprintf("%s <= ?", column), values[0], join, nil
	case OpEq, OpNe:
		return fmt.Sprintf("%s = ?", column), values[0], join, nil
	case OpPt:
		return fmt.Sprintf("(%s IS NULL OR %s = 0)", column, column), nil, join, nil
	case OpNp:
		return fmt.Sprintf("(%s IS NOT NULL AND %s <> 0)", column, column), nil, join, nil
	case OpIn:
		return fmt.Sprintf("%s IN ?", column), values, join, nil
	default:
		return "", nil, "", ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
	}
}

// formatBoolWhere форматирует условие для логических столбцов
func formatBoolWhere(expr ExpressionWhere, column string, join string) (string, interface{}, string, error) {
	var value bool
	if expr.Value == "true" {
		value = true
	} else if expr.Value == "false" {
		value = false
	} else {
		return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column)
	}

	switch expr.Operation {
	case OpNe:
		return fmt.Sprintf("%s = ?", column), value, join, nil
	default:
		return "", nil, "", ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
	}
}

// formatDateTimeWhere форматирует условие для столбцов с типом даты и времени
func formatDateTimeWhere(expr ExpressionWhere, column string, join string) (string, interface{}, string, error) {
	if len(expr.Value) < 3 || expr.Value[0] != '`' || expr.Value[len(expr.Value)-1] != '`' {
		return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column)
	}
	tm, err := time.Parse("2006-01-02 15:04:05", strings.ReplaceAll(expr.Value, "`", ""))
	if err != nil {
		return "", nil, "", ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithError(err)
	}

	value := tm.Format("2006-01-02 15:04:05")

	switch expr.Operation {
	case OpGt:
		return fmt.Sprintf("%s > ?", column), value, join, nil
	case OpLt:
		return fmt.Sprintf("%s < ?", column), value, join, nil
	case OpGte:
		return fmt.Sprintf("%s >= ?", column), value, join, nil
	case OpLte:
		return fmt.Sprintf("%s <= ?", column), value, join, nil
	case OpNe:
		return fmt.Sprintf("%s = ?", column), value, join, nil
	default:
		return "", nil, "", ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
	}
}
