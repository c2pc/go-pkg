package clause

import (
	"github.com/c2pc/go-pkg/v2/utils/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"regexp"
	"strings"
)

func Returning(columns ...string) clause.Returning {
	var returningColumns []clause.Column
	for _, c := range columns {
		if c != "" {
			returningColumns = append(returningColumns, clause.Column{Name: c})
		}
	}

	return clause.Returning{Columns: returningColumns}
}

func Where(query string, args ...interface{}) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if query == "" {
			return tx
		}
		return tx.Where(upperModels(query), args...)
	}
}

func OnConflict(onConflict []interface{}, doUpdates []interface{}) clause.OnConflict {
	var clmns []clause.Column
	var upds []string

	for _, col := range onConflict {
		if v, ok := col.(string); ok {
			clmns = append(clmns, clause.Column{Name: v})
		}
	}

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

func upperModels(model string) string {
	reg := regexp.MustCompile(`"(.)[a-zA-Z]*"[^ ]`)
	return reg.ReplaceAllStringFunc(model, func(w string) string {
		return cases.Title(language.English).String(w)
	})
}

func OrderBy(orderBy map[string]string, tableName string) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		for k, v := range orderBy {
			key := `"` + strings.ReplaceAll(k, ".", `"."`) + `"`
			key = upperModels(key)
			order := model.OrderByAsc
			if strings.ToUpper(v) == model.OrderByDesc {
				order = model.OrderByDesc
			}
			if strings.Index(key, ".") == -1 {
				key = tableName + "." + key
			}
			tx = tx.Order(key + " " + order)
		}
		return tx
	}
}

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
