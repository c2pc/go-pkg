package csv

import (
	"strconv"
	"strings"
)

type IDs []int

func (m *IDs) MarshalCSV() ([]byte, error) {
	var ids []string
	for id := range *m {
		ids = append(ids, strconv.Itoa(id))
	}

	return []byte(strings.Join(ids, ",")), nil
}

func (m *IDs) UnmarshalCSV(data []byte) error {
	var ids []int
	spl := strings.Split(string(data), ",")
	for _, str := range spl {
		id, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	*m = ids
	return nil
}
