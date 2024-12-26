package modelfilter

import (
	"github.com/go-playground/assert/v2"
	"reflect"
	"testing"

	"github.com/c2pc/go-pkg/v2/utils/clause"
)

func TestSortSlice_String(t *testing.T) {
	t.Run("string sort ASC test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "ccc", DBId: "db3"},
			{ID: "aaa", DBId: "db1"},
			{ID: "bbb", DBId: "db2"},
		}

		instructions := []clause.ExpressionOrderBy{
			{Column: "id", Order: clause.OrderByAsc},
		}

		err := SortSlice(calls, instructions, ActiveCallFieldSearchable, getFieldValue)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{
			{ID: "aaa", DBId: "db1"},
			{ID: "bbb", DBId: "db2"},
			{ID: "ccc", DBId: "db3"},
		}, calls)
	})

	t.Run("string sort DESC test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "ccc", DBId: "db3"},
			{ID: "aaa", DBId: "db1"},
			{ID: "bbb", DBId: "db2"},
		}

		instructions := []clause.ExpressionOrderBy{
			{Column: "id", Order: clause.OrderByDesc},
		}

		err := SortSlice(calls, instructions, ActiveCallFieldSearchable, getFieldValue)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{
			{ID: "ccc", DBId: "db3"},
			{ID: "bbb", DBId: "db2"},
			{ID: "aaa", DBId: "db1"},
		}, calls)
	})
}

func TestSortSlice_Int(t *testing.T) {
	t.Run("int sort ASC test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "small", State: 10},
			{ID: "big", State: 100},
			{ID: "middle", State: 50},
		}

		instructions := []clause.ExpressionOrderBy{
			{Column: "state", Order: clause.OrderByAsc},
		}

		err := SortSlice(calls, instructions, ActiveCallFieldSearchable, getFieldValue)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{
			{ID: "small", State: 10},
			{ID: "middle", State: 50},
			{ID: "big", State: 100},
		}, calls)
	})

	t.Run("int sort DESC test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "small", State: 10},
			{ID: "big", State: 100},
			{ID: "middle", State: 50},
		}

		instructions := []clause.ExpressionOrderBy{
			{Column: "state", Order: clause.OrderByDesc},
		}

		err := SortSlice(calls, instructions, ActiveCallFieldSearchable, getFieldValue)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{
			{ID: "big", State: 100},
			{ID: "middle", State: 50},
			{ID: "small", State: 10},
		}, calls)
	})
}

func TestSortSlice_DateTime(t *testing.T) {
	t.Run("datetime sort ASC test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "third", Datestart: "2023-01-03 10:00:00"},
			{ID: "first", Datestart: "2023-01-01 10:00:00"},
			{ID: "second", Datestart: "2023-01-02 10:00:00"},
		}

		instructions := []clause.ExpressionOrderBy{
			{Column: "Datestart", Order: clause.OrderByAsc},
		}

		err := SortSlice(calls, instructions, ActiveCallFieldSearchable, getFieldValue)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, []ActiveCall{{ID: "first", Datestart: "2023-01-01 10:00:00"},
			{ID: "second", Datestart: "2023-01-02 10:00:00"},
			{ID: "third", Datestart: "2023-01-03 10:00:00"}}, calls)
	})

	t.Run("datetime sort DESC test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "third", Datestart: "2023-01-03 10:00:00"},
			{ID: "first", Datestart: "2023-01-01 10:00:00"},
			{ID: "second", Datestart: "2023-01-02 10:00:00"},
		}

		instructions := []clause.ExpressionOrderBy{
			{Column: "Datestart", Order: clause.OrderByDesc},
		}

		err := SortSlice(calls, instructions, ActiveCallFieldSearchable, getFieldValue)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, []ActiveCall{{ID: "third", Datestart: "2023-01-03 10:00:00"},
			{ID: "second", Datestart: "2023-01-02 10:00:00"},
			{ID: "first", Datestart: "2023-01-01 10:00:00"}}, calls)
	})
}

func TestSortSlice_MultiField(t *testing.T) {
	calls := []ActiveCall{
		{ID: "A", DBId: "db1", State: 10},
		{ID: "B", DBId: "db1", State: 20},
		{ID: "C", DBId: "db2", State: 10},
		{ID: "D", DBId: "db2", State: 20},
	}

	instructions := []clause.ExpressionOrderBy{
		{Column: "db_id", Order: clause.OrderByAsc},
		{Column: "state", Order: clause.OrderByDesc},
	}

	err := SortSlice(calls, instructions, ActiveCallFieldSearchable, getFieldValue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotIDs []string
	for _, c := range calls {
		gotIDs = append(gotIDs, c.ID)
	}
	want := []string{"B", "A", "D", "C"}
	if !reflect.DeepEqual(gotIDs, want) {
		t.Errorf("multi-field sort: got %v, want %v", gotIDs, want)
	}
}
