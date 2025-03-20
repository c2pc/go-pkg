package clause

import (
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Returning создает конструкцию для возврата указанных столбцов после выполнения операции
func Returning(columns ...string) clause.Returning {
	var returningColumns []clause.Column
	for _, c := range columns {
		if c != "" {
			returningColumns = append(returningColumns, clause.Column{Name: c})
		}
	}

	return clause.Returning{Columns: returningColumns}
}

// Where создает функцию-замыкание для добавления условия WHERE в запрос GORM
func Where(quoteTo func(string) string, query string, args ...interface{}) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if query == "" {
			return tx
		}
		// Применяем условие WHERE с указанным запросом и аргументами
		return tx.Where(quoteTo(upperModels(query)), args...)
	}
}

// OnConflict создает конструкцию для обработки конфликтов в запросах GORM
func OnConflict(onConflict []interface{}, doUpdates []interface{}) clause.OnConflict {
	var clmns []clause.Column
	var upds []string

	// Добавляем столбцы, по которым нужно проверять конфликты
	for _, col := range onConflict {
		if v, ok := col.(string); ok {
			clmns = append(clmns, clause.Column{Name: v})
		}
	}

	// Добавляем столбцы и значения, которые нужно обновить при конфликте
	for _, upd := range doUpdates {
		if v, ok := upd.(string); ok {
			upds = append(upds, v)
		}
	}

	return clause.OnConflict{
		Columns:   clmns,
		DoUpdates: clause.AssignmentColumns(upds),
	}
}

// upperModels преобразует модель в формат заголовка, используя TitleCase
func upperModels(model string) string {
	reg := regexp.MustCompile(`"([a-zA-Z]+)"[^ ]`)
	return reg.ReplaceAllStringFunc(model, func(w string) string {
		if len(w) > 1 && strings.ToUpper(w[1:2]) == w[1:2] {
			return w
		}
		return cases.Title(language.English).String(w)
	})
}

// Limit создает функцию-замыкание для ограничения количества записей и смещения
func Limit(limit, offset int) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if limit > 0 {
			tx = tx.Limit(limit)
		}
		if offset > 0 {
			tx = tx.Offset(offset)
		}

		return tx
	}
}
