package workererrors

import "fmt"

type WorkerError struct {
	Op    string
	Path  string
	Table string
	Err   error
}

func (e *WorkerError) Error() string {
	return fmt.Sprintf("worker error (%s) [path=%q table=%q]: %v", e.Op, e.Path, e.Table, e.Err)
}

func (e *WorkerError) Unwrap() error {
	return e.Err
}

func NewOpError(op, path, table string, err error) error {
	if err == nil {
		return nil
	}
	return &WorkerError{
		Op:    op,
		Path:  path,
		Table: table,
		Err:   err,
	}
}
