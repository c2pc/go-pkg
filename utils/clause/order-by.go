package clause

import (
	"fmt"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/i18n"
)

var (
	ErrOrderByUnknownColumn = apperr.New("order_by_unknown_column", apperr.WithTextTranslate(i18n.ErrOrderByUnknownColumn), apperr.WithCode(code.InvalidArgument))
)

// FieldOrderBy представляет карту столбцов, по которым можно выполнять сортировку
type FieldOrderBy map[string]Order

// Order описывает информацию о сортировке для столбца
type Order struct {
	Column string // Имя столбца для сортировки
	Join   string // Имя соединения для запроса
}

type ExpressionOrderBy struct {
	Column string `json:"column"` // Имя столбца для сортировки
	Order  string `json:"order"`  // Значение для сортировки
}

// Константы для направлений сортировки
const (
	OrderByAsc  = "ASC"  // Сортировка по возрастанию
	OrderByDesc = "DESC" // Сортировка по убыванию
)

// OrderByFilter строит SQL-запрос для сортировки на основе переданных параметров
func OrderByFilter(quoteTo func(string) string, orderBy []ExpressionOrderBy, fieldOrderBy FieldOrderBy) (string, []string, error) {
	var query []string
	var joins []string

	// Проходим по всем параметрам сортировки
	for _, v := range orderBy {
		if search, ok := fieldOrderBy[v.Column]; ok {
			// Форматируем столбец и соединение
			column := upperModels(quoteTo(search.Column))
			join := quoteTo(search.Join)

			// Определяем направление сортировки
			order := OrderByAsc
			if strings.ToUpper(v.Order) == OrderByDesc {
				order = OrderByDesc
			}

			// Формируем часть запроса для текущего столбца
			query = append(query, fmt.Sprintf("%s %s", column, order))

			// Добавляем соединение, если оно задано
			if join != "" {
				joins = append(joins, join)
			}
		} else {
			// Возвращаем ошибку, если столбец неизвестен
			return "", nil, ErrOrderByUnknownColumn.WithTextArgs(v.Column)
		}
	}

	// Объединяем части запроса и возвращаем результат
	return strings.Join(query, ","), joins, nil
}
