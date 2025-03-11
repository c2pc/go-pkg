package csv

import (
	"strconv"
	"strings"
)

type IntSlice []int

func (m *IntSlice) MarshalCSV() ([]byte, error) {
	var d []string
	for _, n := range *m {
		d = append(d, strconv.Itoa(n))
	}

	if len(d) == 0 {
		return nil, nil
	}

	return []byte(strings.Join(d, ",")), nil
}

func (m *IntSlice) UnmarshalCSV(data []byte) error {
	var d []int
	spl := strings.Split(string(data), ",")
	for _, str := range spl {
		tr := strings.ReplaceAll(str, " ", "")
		if tr == "" {
			continue
		}
		n, err := strconv.Atoi(tr)
		if err != nil {
			return err
		}
		d = append(d, n)
	}
	*m = d
	return nil
}

type StringSlice []string

func (m *StringSlice) MarshalCSV() ([]byte, error) {
	var d []string
	for _, n := range *m {
		d = append(d, n)
	}

	if len(d) == 0 {
		return nil, nil
	}

	return []byte(strings.Join(d, ",")), nil
}

func (m *StringSlice) UnmarshalCSV(data []byte) error {
	var d []string
	spl := strings.Split(string(data), ",")
	for _, str := range spl {
		d = append(d, str)
	}
	*m = d
	return nil
}
