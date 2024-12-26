package modelfilter

import (
	"sort"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/clause"
)

func SortSlice[T any](
	items []T,
	instructions []clause.ExpressionOrderBy,
	searchable clause.FieldSearchable,
	getValueFunc GetFieldValueFunc[T],
) error {
	if len(instructions) == 0 {
		return nil
	}

	type sortKey struct {
		field     string
		asc       bool
		fieldType clause.Type
	}
	var keys []sortKey
	for _, instr := range instructions {
		info, ok := searchable[instr.Column]
		if !ok {
			return clause.ErrOrderByUnknownColumn
		}

		asc := true
		if instr.Order == clause.OrderByDesc {
			asc = false
		}
		keys = append(keys, sortKey{
			field:     instr.Column,
			asc:       asc,
			fieldType: info.Type,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		a := items[i]
		b := items[j]

		for _, key := range keys {
			va, errA := getValueFunc(a, key.field)
			vb, errB := getValueFunc(b, key.field)
			if errA != nil || errB != nil {
				continue
			}

			cmp, err := compareValuesByType(va, vb, key.fieldType)
			if err != nil {
				continue
			}
			if cmp == 0 {
				continue
			}
			if key.asc {
				return cmp < 0
			} else {
				return cmp > 0
			}
		}
		return false
	})

	return nil
}

func compareValuesByType(va, vb interface{}, t clause.Type) (int, error) {
	if va == nil && vb == nil {
		return 0, nil
	}
	if va == nil {
		return -1, nil
	}
	if vb == nil {
		return 1, nil
	}

	switch t {
	case clause.String:
		sa, ok1 := va.(string)
		sb, ok2 := vb.(string)
		if !ok1 || !ok2 {
			return 0, clause.ErrOrderByUnknownColumn
		}
		return strings.Compare(sa, sb), nil

	case clause.Int:
		switch x := va.(type) {
		case int:
			y, ok := vb.(int)
			if !ok {
				return 0, clause.ErrOrderByUnknownColumn
			}
			if x < y {
				return -1, nil
			} else if x > y {
				return 1, nil
			}
			return 0, nil
		case uint64:
			y, ok := vb.(uint64)
			if !ok {
				return 0, clause.ErrOrderByUnknownColumn
			}
			if x < y {
				return -1, nil
			} else if x > y {
				return 1, nil
			}
			return 0, nil
		default:
			return 0, clause.ErrOrderByUnknownColumn
		}

	case clause.Bool:
		ba, ok1 := va.(bool)
		bb, ok2 := vb.(bool)
		if !ok1 || !ok2 {
			return 0, clause.ErrOrderByUnknownColumn
		}
		// false < true
		if ba == bb {
			return 0, nil
		}
		if !ba && bb {
			return -1, nil
		}
		return 1, nil

	case clause.DateTime:
		sa, ok1 := va.(string)
		sb, ok2 := vb.(string)
		if !ok1 || !ok2 {
			return 0, clause.ErrOrderByUnknownColumn
		}
		ta, err := time.Parse("2006-01-02 15:04:05", sa)
		if err != nil {
			return 0, clause.ErrOrderByUnknownColumn
		}
		tb, err := time.Parse("2006-01-02 15:04:05", sb)
		if err != nil {
			return 0, clause.ErrOrderByUnknownColumn
		}
		if ta.Before(tb) {
			return -1, nil
		} else if ta.After(tb) {
			return 1, nil
		}
		return 0, nil

	default:
		return 0, clause.ErrOrderByUnknownColumn
	}
}
