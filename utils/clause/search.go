package clause

import (
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type Condition string

type SearchContext int

const (
	EqualCondition    Condition = "="
	NotEqualCondition Condition = "<>"
	LikeCondition     Condition = "LIKE"
	InCondition       Condition = "IN"
	BetweenCondition  Condition = "BETWEEN"
	QueryCondition    Condition = "QUERY"
)

const (
	AllCtx SearchContext = iota
	WhereCtx
	SearchCtx
)

type FieldSearchable map[string]*Search
type Search struct {
	Condition Condition
	Column    string
	Context   SearchContext
	value     interface{}
}

func SearchFilter(search map[string]interface{}, searchable FieldSearchable) func(tx *gorm.DB) *gorm.DB {
	if len(search) == 0 || len(searchable) == 0 {
		return func(tx *gorm.DB) *gorm.DB { return tx }
	}

	searchData := validateSearchData(search, searchable)

	return func(tx *gorm.DB) *gorm.DB {
		isSearch := false
		db := tx.Session(&gorm.Session{})
		for s, v := range searchData {
			if v != nil {
				if !(v.Context == SearchCtx || v.Context == AllCtx) {
					continue
				}
				column := v.Column
				if column == "" {
					column = s
				}
				switch v.Condition {
				case InCondition:
					if value, ok := v.value.([]interface{}); ok {
						if len(value) > 0 {
							db = db.Or(column+" IN ?", value)
							isSearch = true
						}
					} else if value, ok := v.value.(string); ok {
						if strings.Contains(value, ",") {
							separated := strings.Split(value, ",")
							db = db.Or(column+" IN ?", separated)
							isSearch = true
						} else {
							db = db.Or(column+" = ?", value)
							isSearch = true
						}
					}
				case BetweenCondition:
					if value, ok := v.value.([]interface{}); ok {
						if len(value) == 2 {
							db = db.Or(column+" BETWEEN ? AND ?", value[0], value[1])
							isSearch = true
						}
					} else if value, ok := v.value.(string); ok {
						if strings.Contains(value, ",") {
							separated := strings.Split(value, ",")
							if separated[0] == "" {
								db = db.Or(column+" <= ?", separated[1])
								isSearch = true
							} else if separated[1] == "" {
								db = db.Or(column+" >= ", separated[0])
								isSearch = true
							} else {
								db = db.Or(column+" BETWEEN ? AND ?", separated[0], separated[1])
								isSearch = true
							}
						} else {
							db = db.Or(column+" = ?", value)
							isSearch = true
						}
					}
				case LikeCondition:
					if value, ok := v.value.(string); ok {
						db = db.Or("LOWER("+column+") LIKE LOWER(?)", "%"+value+"%")
						isSearch = true
					} else if value, ok := v.value.(int); ok {
						db = db.Or("CAST("+column+" AS TEXT) LIKE ?", "%"+strconv.Itoa(value)+"%")
						isSearch = true
					} else if value, ok := v.value.(float64); ok {
						db = db.Or("CAST("+column+" AS TEXT) LIKE ?", "%"+strconv.Itoa(int(value))+"%")
						isSearch = true
					}
				case EqualCondition:
					db = db.Or(column+" = ?", v.value)
					isSearch = true
				case NotEqualCondition:
					db = db.Or(column+" <> ?", v.value)
					isSearch = true
				}
			}
		}

		if isSearch {
			tx = tx.Where(db)
		}

		return tx
	}
}

func WhereFilter(search map[string]interface{}, searchable FieldSearchable) func(tx *gorm.DB) *gorm.DB {
	if len(search) == 0 || len(searchable) == 0 {
		return func(tx *gorm.DB) *gorm.DB { return tx }
	}

	searchData := validateSearchData(search, searchable)

	return func(tx *gorm.DB) *gorm.DB {
		for s, v := range searchData {
			if v != nil {
				if !(v.Context == WhereCtx || v.Context == AllCtx) {
					continue
				}
				column := v.Column
				if column == "" {
					column = s
				}
				switch v.Condition {
				case InCondition:
					if value, ok := v.value.([]interface{}); ok {
						if len(value) > 0 {
							tx = tx.Where(column+" IN ?", value)
						}
					} else if value, ok := v.value.(string); ok {
						if strings.Contains(value, ",") {
							separated := strings.Split(value, ",")
							tx = tx.Where(column+" IN ?", separated)
						} else {
							tx = tx.Where(column+" = ?", value)
						}
					}
				case BetweenCondition:
					if value, ok := v.value.([]interface{}); ok {
						if len(value) == 2 {
							tx = tx.Where(column+" BETWEEN ? AND ?", value[0], value[1])
						}
					} else if value, ok := v.value.(string); ok {
						if strings.Contains(value, ",") {
							separated := strings.Split(value, ",")
							if separated[0] == "" {
								tx = tx.Where(column+" <= ?", separated[1])
							} else if separated[1] == "" {
								tx = tx.Where(column+" >= ", separated[0])
							} else {
								tx = tx.Where(column+" BETWEEN ? AND ?", separated[0], separated[1])
							}
						} else {
							tx = tx.Where(column+" = ?", value)
						}
					}
				case LikeCondition:
					if value, ok := v.value.(string); ok {
						tx = tx.Where("LOWER("+column+") LIKE LOWER(?)", "%"+value+"%")
					} else if value, ok := v.value.(int); ok {
						tx = tx.Where("CAST("+column+" AS TEXT) LIKE ?", "%"+strconv.Itoa(value)+"%")
					} else if value, ok := v.value.(float64); ok {
						tx = tx.Where("CAST("+column+" AS TEXT) LIKE ?", "%"+strconv.Itoa(int(value))+"%")
					}
				case EqualCondition:
					tx = tx.Where(column+" = ?", v.value)
				case NotEqualCondition:
					tx = tx.Where(column+" <> ?", v.value)
				}
			}
		}
		return tx
	}
}

func validateSearchData(searchData map[string]interface{}, searchable FieldSearchable) FieldSearchable {
	validSearch := FieldSearchable{}
	for field, value := range searchData {
		if _, ok := searchable[field]; ok {
			if value != "" && field != "" {
				searchField := &Search{}
				searchField.Condition = searchable[field].Condition
				searchField.Column = searchable[field].Column
				searchField.Context = searchable[field].Context
				searchField.value = value
				validSearch[field] = searchField
			}
		}
	}

	return validSearch
}
