package modelfilter

import (
	"fmt"
	"github.com/go-playground/assert/v2"
	"reflect"
	"testing"

	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/request"
)

type ActiveCall struct {
	ID             string  `json:"id"`
	DBId           string  `json:"db_id"`
	State          int     `json:"state"`
	UserFlags      string  `json:"user_flags"`
	Timestart      uint64  `json:"timestart"`
	Datestart      string  `json:"datestart"`
	Timeout        uint64  `json:"timeout"`
	Dateout        string  `json:"dateout"`
	Callid         *string `json:"callid"`
	FromUri        string  `json:"from_uri"`
	ToUri          string  `json:"to_uri"`
	CallerTag      string  `json:"caller_tag"`
	CallerContact  string  `json:"caller_contact"`
	CalleeCseq     string  `json:"callee_cseq"`
	CallerRouteSet string  `json:"caller_route_set"`
	CallerBindAddr string  `json:"caller_bind_addr"`
	CallerSdp      string  `json:"caller_sdp"`
	CallerSentSdp  string  `json:"caller_sent_sdp"`
	GGG            struct {
		HHH int `json:"hhh"`
	} `json:"ggg"`
}

var ActiveCallFieldSearchable = clause.FieldSearchable{
	"id":             {Column: "id", Type: clause.String},
	"db_id":          {Column: "db_id", Type: clause.String},
	"ggg.hhh":        {Column: "ggg.hhh", Type: clause.Int},
	"state":          {Column: "state", Type: clause.Int},
	"UserFlags":      {Column: "user_flags", Type: clause.String},
	"Timestart":      {Column: "timestart", Type: clause.Int},
	"Datestart":      {Column: "Datestart", Type: clause.DateTime},
	"Timeout":        {Column: "timeout", Type: clause.Int},
	"Dateout":        {Column: "dateout", Type: clause.DateTime},
	"Callid":         {Column: "callid", Type: clause.String},
	"FromUri":        {Column: "from_uri", Type: clause.String},
	"ToUri":          {Column: "to_uri", Type: clause.String},
	"CallerTag":      {Column: "caller_tag", Type: clause.String},
	"CallerContact":  {Column: "caller_contact", Type: clause.String},
	"CalleeCseq":     {Column: "callee_cseq", Type: clause.String},
	"CallerRouteSet": {Column: "caller_route_set", Type: clause.String},
	"CallerBindAddr": {Column: "caller_bind_addr", Type: clause.String},
	"CallerSdp":      {Column: "caller_sdp", Type: clause.String},
	"CallerSentSdp":  {Column: "caller_sent_sdp", Type: clause.String},
}

func getFieldValue(call ActiveCall, field string) (interface{}, error) {
	switch field {
	case "id":
		return call.ID, nil
	case "db_id", "DBId":
		return call.DBId, nil
	case "state", "State":
		return call.State, nil
	case "UserFlags":
		return call.UserFlags, nil
	case "Timestart":
		return call.Timestart, nil
	case "Datestart":
		return call.Datestart, nil
	case "Timeout":
		return call.Timeout, nil
	case "Dateout":
		return call.Dateout, nil
	case "Callid":
		return call.Callid, nil
	case "FromUri":
		return call.FromUri, nil
	case "ToUri":
		return call.ToUri, nil
	case "CallerTag":
		return call.CallerTag, nil
	case "CallerContact":
		return call.CallerContact, nil
	case "CalleeCseq":
		return call.CalleeCseq, nil
	case "CallerRouteSet":
		return call.CallerRouteSet, nil
	case "CallerBindAddr":
		return call.CallerBindAddr, nil
	case "CallerSdp":
		return call.CallerSdp, nil
	case "CallerSentSdp":
		return call.CallerSentSdp, nil
	case "ggg.hhh":
		return call.GGG.HHH, nil
	default:
		return nil, clause.ErrFilterUnknownColumn
	}
}

func TestApplyFilters_String(t *testing.T) {
	t.Run("eq string test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "aaa", DBId: "db1", State: 10, UserFlags: "foo"},
			{ID: "bbb", DBId: "db2", State: 20, UserFlags: "bar"},
			{ID: "ccc", DBId: "db2", State: 30, UserFlags: "baz"},
		}
		expr, err := request.ParseWhere("UserFlags eq bar")
		exprs := append([]clause.ExpressionWhere{}, *expr)
		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "bbb", DBId: "db2", State: 20, UserFlags: "bar"}}, got)
	})

	t.Run("ne string test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "aaa", DBId: "db1", State: 10, UserFlags: "foo"},
			{ID: "bbb", DBId: "db2", State: 20, UserFlags: "bar"},
			{ID: "ccc", DBId: "db2", State: 30, UserFlags: "baz"},
		}
		expr, err := request.ParseWhere("UserFlags = bar")
		exprs := append([]clause.ExpressionWhere{}, *expr)
		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "bbb", DBId: "db2", State: 20, UserFlags: "bar"}}, got)
	})

	t.Run("co string test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "aaa", DBId: "db1", State: 10, UserFlags: "foo"},
			{ID: "bbb", DBId: "db2", State: 20, UserFlags: "bar"},
			{ID: "ccc", DBId: "db2", State: 30, UserFlags: "baz"},
		}
		expr, err := request.ParseWhere("UserFlags co ba")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		exprs := append([]clause.ExpressionWhere{}, *expr)
		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "bbb", DBId: "db2", State: 20, UserFlags: "bar"}, {ID: "ccc", DBId: "db2", State: 30, UserFlags: "baz"}}, got)
	})

	t.Run("sw string test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "aaa", DBId: "db1", State: 10, UserFlags: "foo"},
			{ID: "bbb", DBId: "db2", State: 20, UserFlags: "bar"},
			{ID: "ccc", DBId: "db3", State: 30, UserFlags: "baz"},
			{ID: "ggg", DBId: "db4", State: 40, UserFlags: "faz"},
		}
		expr, err := request.ParseWhere("UserFlags sw f")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		exprs := append([]clause.ExpressionWhere{}, *expr)
		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "aaa", DBId: "db1", State: 10, UserFlags: "foo"}, {ID: "ggg", DBId: "db4", State: 40, UserFlags: "faz"}}, got)
	})

	t.Run("ew string test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "aaa", DBId: "db1", State: 10, UserFlags: "foo"},
			{ID: "bbb", DBId: "db2", State: 20, UserFlags: "bar"},
			{ID: "ccc", DBId: "db3", State: 30, UserFlags: "baz"},
			{ID: "ggg", DBId: "db4", State: 40, UserFlags: "faz"},
		}
		expr, err := request.ParseWhere("UserFlags ew z")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		exprs := append([]clause.ExpressionWhere{}, *expr)
		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "ccc", DBId: "db3", State: 30, UserFlags: "baz"}, {ID: "ggg", DBId: "db4", State: 40, UserFlags: "faz"}}, got)
	})

	t.Run("in string test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "aaa", DBId: "db1", State: 10, UserFlags: "A"},
			{ID: "bbb", DBId: "db2", State: 20, UserFlags: "bar"},
			{ID: "ccc", DBId: "db3", State: 30, UserFlags: "baz"},
			{ID: "ggg", DBId: "db4", State: 40, UserFlags: "baz"},
		}
		expr, err := request.ParseWhere("UserFlags in `bar,A`")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		fmt.Printf("%+v", expr)

		exprs := append([]clause.ExpressionWhere{}, *expr)
		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "aaa", DBId: "db1", State: 10, UserFlags: "A"}, {ID: "bbb", DBId: "db2", State: 20, UserFlags: "bar"}}, got)
	})

	t.Run("nne string test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "aaa", DBId: "db1", State: 10, UserFlags: "A"},
			{ID: "bbb", DBId: "db2", State: 20, UserFlags: "bar"},
			{ID: "ccc", DBId: "db3", State: 30, UserFlags: "baz"},
			{ID: "ggg", DBId: "db4", State: 40, UserFlags: "baz"},
		}
		expr, err := request.ParseWhere("UserFlags <> baz")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		fmt.Printf("%+v", expr)

		exprs := append([]clause.ExpressionWhere{}, *expr)
		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "aaa", DBId: "db1", State: 10, UserFlags: "A"}, {ID: "bbb", DBId: "db2", State: 20, UserFlags: "bar"}}, got)
	})

}

func TestApplyFilters_Int(t *testing.T) {
	t.Run("> int test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "A", State: 10},
			{ID: "B", State: 15},
			{ID: "C", State: 20},
		}

		expr, err := request.ParseWhere("state > 10")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, []ActiveCall{{ID: "B", State: 15}, {ID: "C", State: 20}}, got)
	})

	t.Run(">= int test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "A", State: 10},
			{ID: "B", State: 15},
			{ID: "C", State: 20},
		}

		expr, err := request.ParseWhere("state >= 15")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, []ActiveCall{{ID: "B", State: 15}, {ID: "C", State: 20}}, got)
	})

	t.Run("< int test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "A", State: 10},
			{ID: "B", State: 15},
			{ID: "C", State: 20},
		}

		expr, err := request.ParseWhere("state < 15")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, []ActiveCall{{ID: "A", State: 10}}, got)
	})

	t.Run("<= int test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "A", State: 10},
			{ID: "B", State: 15},
			{ID: "C", State: 20},
		}

		expr, err := request.ParseWhere("state <= 15")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, []ActiveCall{{ID: "A", State: 10}, {ID: "B", State: 15}}, got)
	})

	t.Run("eq int test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "A", State: 10},
			{ID: "B", State: 15},
			{ID: "C", State: 20},
		}

		expr, err := request.ParseWhere("state eq 15")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, []ActiveCall{{ID: "B", State: 15}}, got)
	})

	t.Run("= int test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "A", State: 10},
			{ID: "B", State: 15},
			{ID: "C", State: 20},
		}

		expr, err := request.ParseWhere("state = 15")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, []ActiveCall{{ID: "B", State: 15}}, got)
	})

	t.Run("in int test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "A", State: 10},
			{ID: "B", State: 15},
			{ID: "C", State: 20},
		}

		expr, err := request.ParseWhere("state in `15,10`")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, []ActiveCall{{ID: "A", State: 10}, {ID: "B", State: 15}}, got)
	})

	t.Run("nin int test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "A", State: 10},
			{ID: "B", State: 15},
			{ID: "C", State: 20},
		}

		expr, err := request.ParseWhere("state nin `10`")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, []ActiveCall{{ID: "B", State: 15}, {ID: "C", State: 20}}, got)
	})

	t.Run("nne int test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "A", State: 10},
			{ID: "B", State: 15},
			{ID: "C", State: 20},
		}

		expr, err := request.ParseWhere("state <> 15")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, []ActiveCall{{ID: "A", State: 10}, {ID: "C", State: 20}}, got)
	})
}

func TestApplyFilters_DateTime(t *testing.T) {
	t.Run(">= datetime test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "1", Datestart: "2023-01-01 10:00:00"},
			{ID: "2", Datestart: "2023-01-02 15:00:00"},
			{ID: "3", Datestart: "2023-01-03 15:00:00"},
			{ID: "4", Datestart: "2023-01-03 16:00:00"},
			{ID: "5", Datestart: "2023-01-03 16:02:00"},
			{ID: "6", Datestart: "2023-01-03 20:00:00"},
		}
		expr, err := request.ParseWhere("Datestart >= `2023-01-03 16:00:00`")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "4", Datestart: "2023-01-03 16:00:00"}, {ID: "5", Datestart: "2023-01-03 16:02:00"}, {ID: "6", Datestart: "2023-01-03 20:00:00"}}, got)
	})

	t.Run("> datetime test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "1", Datestart: "2023-01-01 10:00:00"},
			{ID: "2", Datestart: "2023-01-02 15:00:00"},
			{ID: "3", Datestart: "2023-01-03 15:00:00"},
			{ID: "4", Datestart: "2023-01-03 16:00:00"},
			{ID: "5", Datestart: "2023-01-03 16:02:00"},
			{ID: "6", Datestart: "2023-01-03 20:00:00"},
		}
		expr, err := request.ParseWhere("Datestart > `2023-01-03 16:00:00`")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "5", Datestart: "2023-01-03 16:02:00"}, {ID: "6", Datestart: "2023-01-03 20:00:00"}}, got)
	})

	t.Run("< datetime test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "1", Datestart: "2023-01-01 10:00:00"},
			{ID: "2", Datestart: "2023-01-02 15:00:00"},
			{ID: "3", Datestart: "2023-01-03 15:00:00"},
			{ID: "4", Datestart: "2023-01-03 16:00:00"},
			{ID: "5", Datestart: "2023-01-03 16:02:00"},
			{ID: "6", Datestart: "2023-01-03 20:00:00"},
		}
		expr, err := request.ParseWhere("Datestart < `2023-01-03 16:00:00`")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "1", Datestart: "2023-01-01 10:00:00"}, {ID: "2", Datestart: "2023-01-02 15:00:00"}, {ID: "3", Datestart: "2023-01-03 15:00:00"}}, got)
	})

	t.Run("<= datetime test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "1", Datestart: "2023-01-01 10:00:00"},
			{ID: "2", Datestart: "2023-01-02 15:00:00"},
			{ID: "3", Datestart: "2023-01-03 15:00:00"},
			{ID: "4", Datestart: "2023-01-03 16:00:00"},
			{ID: "5", Datestart: "2023-01-03 16:02:00"},
			{ID: "6", Datestart: "2023-01-03 20:00:00"},
		}
		expr, err := request.ParseWhere("Datestart <= `2023-01-03 16:00:00`")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "1", Datestart: "2023-01-01 10:00:00"}, {ID: "2", Datestart: "2023-01-02 15:00:00"}, {ID: "3", Datestart: "2023-01-03 15:00:00"}, {ID: "4", Datestart: "2023-01-03 16:00:00"}}, got)
	})

	t.Run("eq datetime test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "1", Datestart: "2023-01-01 10:00:00"},
			{ID: "2", Datestart: "2023-01-02 15:00:00"},
			{ID: "3", Datestart: "2023-01-03 15:00:00"},
			{ID: "4", Datestart: "2023-01-03 16:00:00"},
			{ID: "5", Datestart: "2023-01-03 16:02:00"},
			{ID: "6", Datestart: "2023-01-03 20:00:00"},
		}
		expr, err := request.ParseWhere("Datestart eq `2023-01-03 16:00:00`")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "4", Datestart: "2023-01-03 16:00:00"}}, got)
	})

	t.Run("= datetime test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "1", Datestart: "2023-01-01 10:00:00"},
			{ID: "2", Datestart: "2023-01-02 15:00:00"},
			{ID: "3", Datestart: "2023-01-03 15:00:00"},
			{ID: "4", Datestart: "2023-01-03 16:00:00"},
			{ID: "5", Datestart: "2023-01-03 16:02:00"},
			{ID: "6", Datestart: "2023-01-03 20:00:00"},
		}
		expr, err := request.ParseWhere("Datestart = `2023-01-03 16:00:00`")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "4", Datestart: "2023-01-03 16:00:00"}}, got)
	})

	t.Run("<> datetime test", func(t *testing.T) {
		calls := []ActiveCall{
			{ID: "1", Datestart: "2023-01-01 10:00:00"},
			{ID: "2", Datestart: "2023-01-02 15:00:00"},
			{ID: "3", Datestart: "2023-01-03 15:00:00"},
			{ID: "4", Datestart: "2023-01-03 16:00:00"},
			{ID: "5", Datestart: "2023-01-03 16:02:00"},
			{ID: "6", Datestart: "2023-01-03 20:00:00"},
		}
		expr, err := request.ParseWhere("Datestart <> `2023-01-03 16:00:00`")
		fmt.Printf("%+v", expr)
		exprs := append([]clause.ExpressionWhere{}, *expr)

		got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Equal(t, []ActiveCall{{ID: "1", Datestart: "2023-01-01 10:00:00"},
			{ID: "2", Datestart: "2023-01-02 15:00:00"},
			{ID: "3", Datestart: "2023-01-03 15:00:00"},
			{ID: "5", Datestart: "2023-01-03 16:02:00"},
			{ID: "6", Datestart: "2023-01-03 20:00:00"},
		}, got)
	})

}

func TestApplyFilters_ComplexExpressions(t *testing.T) {
	calls := []ActiveCall{
		{ID: "1", DBId: "db1", State: 10, Datestart: "2023-06-30 12:00:00"},
		{ID: "2", DBId: "db2", State: 20, Datestart: "2023-06-30 13:00:00"},
		{ID: "3", DBId: "db3", State: 30, Datestart: "2023-06-30 13:00:00"},
		{ID: "4", DBId: "db4", State: 40, Datestart: "2023-06-30 14:00:00"},
		{ID: "5", DBId: "db5", State: 50, Datestart: "2023-06-30 15:00:00"},
		{ID: "6", DBId: "db6", State: 70, Datestart: "2023-08-30 16:00:00"},
		{ID: "7", DBId: "db7", State: 120, Datestart: "2023-12-30 17:00:00"},
		{ID: "8", DBId: "db8", State: 100, Datestart: "2023-02-20 18:00:00"},
		{ID: "9", DBId: "db9", State: 1, Datestart: "2023-08-30 20:00:00"},
	}
	expr, err := request.ParseWhere("Callid pt and            ((id eq 1 or id eq 2 or id eq 3 or id eq 4 or id eq 5) and (state > 10 and Datestart <= `2023-06-30 15:00:00`)) or state = 1")
	fmt.Printf("%+v \n", expr.Expressions)
	exprs := append([]clause.ExpressionWhere{}, *expr)

	got, err := ApplyFilters(calls, ActiveCallFieldSearchable, getFieldValue, exprs)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}

	var gotIDs []string
	for _, c := range got {
		gotIDs = append(gotIDs, c.ID)
	}

	wantIDs := []string{"2", "3", "4", "5", "9"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Errorf("ожидалось %v, получили %v", wantIDs, gotIDs)
	}
}
