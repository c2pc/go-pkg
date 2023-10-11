package datetime

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"
)

const TimeFormat = "2006-01-02 15:04:05"

type DateTime time.Time

func (t *DateTime) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 2 {
		*t = DateTime(time.Time{})
		return
	}

	now, err := time.Parse(`"`+TimeFormat+`"`, string(data))
	*t = DateTime(now)
	return
}

func (t DateTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(TimeFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, TimeFormat)
	b = append(b, '"')
	return b, nil
}

func (t DateTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	tlt := time.Time(t)

	if tlt.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return tlt.Format(TimeFormat), nil
}

func (t *DateTime) Scan(v interface{}) error {
	if value, ok := v.(time.Time); ok {
		*t = DateTime(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

func (t DateTime) String() string {
	return time.Time(t).Format(TimeFormat)
}

func ValidateJSONDateType(field reflect.Value) interface{} {
	if field.Type() == reflect.TypeOf(DateTime{}) {
		timeStr := field.Interface().(DateTime).String()
		if timeStr == "0001-01-01 00:00:00" {
			return nil
		}
		return timeStr
	}
	return nil
}
