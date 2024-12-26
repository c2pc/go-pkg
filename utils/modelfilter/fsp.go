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
	objs, err := ApplyFilters[T](objs, searchable, FilterFunc, m.Where)
	if err != nil {
		return err
	}

	if m.MustReturnTotalRows {
		m.TotalRows = int64(len(objs))
	}
	if m.OrderBy != nil {
		err = SortSlice[T](objs, m.OrderBy, searchable, SorterFunc)
		if err != nil {
			return err
		}
	}

	m.Rows = ApplyLimits[T](objs, m.Offset, m.Limit)

	return nil
}
