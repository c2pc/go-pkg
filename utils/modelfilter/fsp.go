package modelfilter

import (
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/model"
)

func FSP[T any](objs []T,
	searchable clause.FieldSearchable,
	FilterFunc GetFieldValueFunc[T],
	SorterFunc GetFieldValueFunc[T],
	m model.Meta[T],
) error {
	filterObjs, err := ApplyFilters[T](objs, searchable, FilterFunc, m.Where.Expressions)
	if err != nil {
		return err
	}

	if m.MustReturnTotalRows {
		m.TotalRows = int64(len(filterObjs))
	}

	err = SortSlice[T](filterObjs, m.OrderBy, searchable, SorterFunc)
	if err != nil {
		return err
	}

	limitobjs := ApplyLimits[T](filterObjs, m.Offset, m.Limit)

	m.Rows = limitobjs

	return nil
}
