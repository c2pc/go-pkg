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

const (
	TypeDateTime = "2006-01-02 15:04:05"
	TypeTime     = "15:04:05"
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
	Time     Type = "time"     // Дата и время
)

// Константы операций фильтрации
const (
	OpIn  = "in"  // Операция "IN"
	OpNin = "nin" // Операция "NOT IN"
	OpPt  = "pt"  // Операция "IS NULL OR = ''"
	OpNp  = "np"  // Операция "IS NOT NULL AND <> ''"
	OpCo  = "co"  // Операция "LIKE %...%"
	OpEq  = "eq"  // Операция "="
	OpSw  = "sw"  // Операция "LIKE ...%"
	OpEw  = "ew"  // Операция "LIKE %..."
	OpGt  = ">"   // Операция ">"
	OpLt  = "<"   // Операция "<"
	OpGte = ">="  // Операция ">="
	OpLte = "<="  // Операция "<="
	OpNe  = "="   // Операция "="
	OpNne = "!="  // Операция "<>"
)

var Operators = []string{OpIn, OpNin, OpPt, OpNp, OpCo, OpEq, OpSw, OpEw, OpGt, OpLt, OpGte, OpLte, OpNe, OpNne}

// FieldSearchable представляет карту столбцов с информацией о том, как они можно искать
type FieldSearchable map[string]Search

// Search содержит информацию о том, как искать в определенном столбце
type Search struct {
	Column string // Имя столбца в базе данных
	Type   Type   // Тип данных в столбце
	Join   string // Имя соединения для запроса
	SQL    string // SQL запрос
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
			if search, ok := fieldSearchable[expr.Column]; ok {
				query, args, join, splitArgs, err := formatWhereString(quoteTo, expr, search)
				if err != nil {
					return "", nil, nil, err
				}
				result = append(result, query)
				if args != nil {
					//Добавляем элементы
					if search.SQL != "" {
						if t, is := args.([]interface{}); is {
							values = append(values, t...)
						} else {
							if a, ok := args.([]interface{}); ok && splitArgs {
								values = append(values, a...)
							} else {
								values = append(values, args)
							}
						}
					} else {
						if a, ok := args.([]interface{}); ok && splitArgs {
							values = append(values, a...)
						} else {
							values = append(values, args)
						}
					}
				}
				if join != "" {
					joins = append(joins, join)
				}
			} else {
				return "", nil, nil, ErrFilterUnknownColumn.WithTextArgs(expr.Column)
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
func formatWhereString(quoteTo func(string) string, expr ExpressionWhere, search Search) (string, interface{}, string, bool, error) {
	column := upperModels(quoteTo(search.Column))
	join := quoteTo(search.Join)

	query, args, splitArgs, err := func() (string, interface{}, bool, error) {
		if search.SQL != "" {
			return formatSQLWhere(expr, search.SQL)
		}
		switch search.Type {
		case String:
			return formatStringWhere(expr, column)
		case Int:
			return formatIntWhere(expr, column)
		case Bool:
			return formatBoolWhere(expr, column)
		case DateTime:
			return formatDateTimeWhere(expr, column)
		case Time:
			return formatTimeWhere(expr, column)
		default:
			return "", nil, false, fmt.Errorf("unknown type %s for %s", search.Type, expr.Column)
		}
	}()
	if err != nil {
		return "", nil, "", false, err
	}

	return query, args, join, splitArgs, nil
}

// formatStringWhere форматирует условие для строковых столбцов
func formatStringWhere(expr ExpressionWhere, column string) (string, interface{}, bool, error) {
	var values []string
	if expr.Operation != OpIn && expr.Operation != OpNin && expr.Operation != OpPt && expr.Operation != OpNp {
		if len(expr.Value) < 3 || expr.Value[0] != '`' || expr.Value[len(expr.Value)-1] != '`' {
			return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithErrorText("string len(expr.Value) < 3")
		}
		val := strings.ReplaceAll(expr.Value, "`", "")
		values = []string{val}
	} else if expr.Operation == OpIn || expr.Operation == OpNin {
		values = []string{}
		vals := strings.Split(expr.Value, "`,`")
		for _, val := range vals {
			val = strings.ReplaceAll(val, "`", "")
			values = append(values, val)
		}
		if len(values) == 0 {
			return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithErrorText("string len(values) == 0")
		}
	}

	switch expr.Operation {
	case OpEq:
		return fmt.Sprintf("%s = ?", column), values[0], false, nil
	case OpCo:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", column), "%" + values[0] + "%", false, nil
	case OpSw:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", column), values[0] + "%", false, nil
	case OpEw:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", column), "%" + values[0], false, nil
	case OpPt:
		return fmt.Sprintf("(%s IS NULL OR %s = '')", column, column), nil, false, nil
	case OpNp:
		return fmt.Sprintf("(%s IS NOT NULL AND %s <> '')", column, column), nil, false, nil
	case OpIn:
		return fmt.Sprintf("%s IN ?", column), values, false, nil
	case OpNin:
		return fmt.Sprintf("%s NOT IN ?", column), values, false, nil
	default:
		return "", nil, false, ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
	}
}

// formatIntWhere форматирует условие для целочисленных столбцов
func formatIntWhere(expr ExpressionWhere, column string) (string, interface{}, bool, error) {
	var values []int
	if expr.Operation != OpIn && expr.Operation != OpNin && expr.Operation != OpPt && expr.Operation != OpNp {
		val, err := strconv.Atoi(expr.Value)
		if err != nil {
			return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithError(err)
		}
		values = []int{val}
	} else if expr.Operation == OpIn || expr.Operation == OpNin {
		values = []int{}
		vals := strings.Split(expr.Value, ",")
		for _, val := range vals {
			v, err := strconv.Atoi(val)
			if err != nil {
				return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithError(err)
			}
			values = append(values, v)
		}
		if len(values) == 0 {
			return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithErrorText("int len(values) == 0")
		}
	}

	switch expr.Operation {
	case OpGt:
		return fmt.Sprintf("%s > ?", column), values[0], false, nil
	case OpLt:
		return fmt.Sprintf("%s < ?", column), values[0], false, nil
	case OpGte:
		return fmt.Sprintf("%s >= ?", column), values[0], false, nil
	case OpLte:
		return fmt.Sprintf("%s <= ?", column), values[0], false, nil
	case OpEq, OpNe:
		return fmt.Sprintf("%s = ?", column), values[0], false, nil
	case OpNne:
		return fmt.Sprintf("%s <> ?", column), values[0], false, nil
	case OpPt:
		return fmt.Sprintf("(%s IS NULL OR %s = 0)", column, column), nil, false, nil
	case OpNp:
		return fmt.Sprintf("(%s IS NOT NULL AND %s <> 0)", column, column), nil, false, nil
	case OpIn:
		return fmt.Sprintf("%s IN ?", column), values, false, nil
	case OpNin:
		return fmt.Sprintf("%s NOT IN ?", column), values, false, nil
	default:
		return "", nil, false, ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
	}
}

// formatBoolWhere форматирует условие для логических столбцов
func formatBoolWhere(expr ExpressionWhere, column string) (string, interface{}, bool, error) {
	var values []bool
	if expr.Operation != OpIn && expr.Operation != OpNin {
		if expr.Value == "true" {
			values = []bool{true}
		} else if expr.Value == "false" {
			values = []bool{false}
		} else {
			return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithErrorText("bool not true or false")
		}
	} else {
		values = []bool{}
		vals := strings.Split(expr.Value, ",")
		for _, val := range vals {
			if val == "true" {
				values = append(values, true)
			} else if val == "false" {
				values = append(values, false)
			} else {
				return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithErrorText("bool not true or false")
			}
		}
		if len(values) == 0 {
			return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithErrorText("int len(values) == 0")
		}
	}

	switch expr.Operation {
	case OpNe:
		return fmt.Sprintf("%s = ?", column), values[0], false, nil
	case OpIn:
		return fmt.Sprintf("%s IN ?", column), values, false, nil
	case OpNin:
		return fmt.Sprintf("%s NOT IN ?", column), values, false, nil
	default:
		return "", nil, false, ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
	}
}

// formatDateTimeWhere форматирует условие для столбцов с типом даты и времени
func formatDateTimeWhere(expr ExpressionWhere, column string) (string, interface{}, bool, error) {
	if len(expr.Value) < 3 || expr.Value[0] != '`' || expr.Value[len(expr.Value)-1] != '`' {
		return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithErrorText("datetime len(expr.Value) < 3")
	}
	tm, err := time.Parse(TypeDateTime, strings.ReplaceAll(expr.Value, "`", ""))
	if err != nil {
		return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithError(err)
	}

	value := tm.Format(TypeDateTime)

	switch expr.Operation {
	case OpGt:
		return fmt.Sprintf("%s > ?", column), value, false, nil
	case OpLt:
		return fmt.Sprintf("%s < ?", column), value, false, nil
	case OpGte:
		return fmt.Sprintf("%s >= ?", column), value, false, nil
	case OpLte:
		return fmt.Sprintf("%s <= ?", column), value, false, nil
	case OpNe:
		return fmt.Sprintf("%s >= ? AND %s <= ?", column, column), []interface{}{value, value}, true, nil
	default:
		return "", nil, false, ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
	}
}

// formatTimeWhere форматирует условие для столбцов с типом времени
func formatTimeWhere(expr ExpressionWhere, column string) (string, interface{}, bool, error) {
	if len(expr.Value) < 3 || expr.Value[0] != '`' || expr.Value[len(expr.Value)-1] != '`' {
		return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithErrorText("time len(expr.Value) < 3")
	}
	tm, err := time.Parse(TypeTime, strings.ReplaceAll(expr.Value, "`", ""))
	if err != nil {
		return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithError(err)
	}

	value := tm.Format(TypeTime)

	switch expr.Operation {
	case OpGt:
		return fmt.Sprintf("%s > ?", column), value, false, nil
	case OpLt:
		return fmt.Sprintf("%s < ?", column), value, false, nil
	case OpGte:
		return fmt.Sprintf("%s >= ?", column), value, false, nil
	case OpLte:
		return fmt.Sprintf("%s <= ?", column), value, false, nil
	case OpNe:
		return fmt.Sprintf("%s >= ? AND %s <= ?", column, column), []interface{}{value, value}, true, nil
	default:
		return "", nil, false, ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
	}
}

// formatDateTimeWhere форматирует условие для столбцов с типом даты и времени
func formatSQLWhere(expr ExpressionWhere, column string) (string, interface{}, bool, error) {
	if expr.Operation != OpIn {
		return "", nil, false, ErrFilterUnknownOperator.WithTextArgs(expr.Operation, expr.Column)
	}

	_, arg, _, err := formatIntWhere(expr, "")
	if err != nil {
		return "", nil, false, err
	}

	var args []interface{}
	if v, ok := arg.([]int); ok {
		for _, s := range v {
			args = append(args, s)
		}
	} else {
		return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithErrorText("sql not interface []int")
	}

	qCount := strings.Count(column, "?")
	if qCount != len(args) {
		return "", nil, false, ErrFilterInvalidValue.WithTextArgs(expr.Value, expr.Column).WithErrorText("sql qCount != len(args)")
	}

	return column, args, false, nil
}
