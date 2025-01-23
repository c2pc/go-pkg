package udatetime

import (
	"time"
)

func ConvertToUTC(t time.Time) time.Time {

	if _, off := t.Zone(); off != 0 {
		return t.UTC()
	}

	localLocation := time.Now().Location()

	localTime := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		localLocation,
	)
	return localTime.UTC()
}
