package task_test

import (
	"context"
	"errors"
	"testing"

	"github.com/c2pc/go-pkg/v2/task"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/stretchr/testify/assert"
)

func TestMassDelete(t *testing.T) {
	ctx := context.Background()
	type Input struct {
		IDs []int `json:"ids"`
	}

	data := []byte(`{"ids": [1, 2, 3]}`)
	checkDatafn := func(input Input) error { return nil }
	idsFn := func(input Input) []int {
		return input.IDs
	}

	pluckIDsFn := func(ctx context.Context, ids []int) ([]int, error) {
		return []int{2, 3}, nil
	}
	actionFn := func(ctx context.Context, input Input, id int) error {
		if id == 3 {
			return apperr.ErrBadRequest
		}
		return nil
	}

	notFoundError := apperr.ErrNotFound

	msg, err := task.MassDelete(ctx, data, notFoundError, checkDatafn, idsFn, pluckIDsFn, actionFn)

	assert.NoError(t, err)
	assert.Equal(t, 3, msg.GetCount())
	assert.Len(t, msg.GetErrors(), 2)
	assert.Len(t, msg.GetSuccesses(), 1)
	assert.Equal(t, "1", msg.GetErrors()[0].Key)
	assert.Equal(t, "Not found error", msg.GetErrors()[0].Value)
	assert.Equal(t, "3", msg.GetErrors()[1].Key)
	assert.Equal(t, "Bad request error", msg.GetErrors()[1].Value)
}

func TestMassUpdate(t *testing.T) {
	ctx := context.Background()
	type Input struct {
		IDs []int `json:"ids"`
	}

	data := []byte(`{"ids": [1, 2, 3]}`)
	checkDatafn := func(input Input) error { return nil }
	idsFn := func(input Input) []int {
		return input.IDs
	}

	pluckIDsFn := func(ctx context.Context, ids []int) ([]int, error) {
		return []int{2, 3}, nil
	}
	actionFn := func(ctx context.Context, id int, input Input) error {
		if id == 3 {
			return apperr.ErrBadRequest
		}
		return nil
	}

	notFoundError := apperr.ErrNotFound

	msg, err := task.MassUpdate(ctx, data, notFoundError, checkDatafn, idsFn, pluckIDsFn, actionFn)

	assert.NoError(t, err)
	assert.Equal(t, 3, msg.GetCount())
	assert.Len(t, msg.GetErrors(), 2)
	assert.Len(t, msg.GetSuccesses(), 1)
	assert.Equal(t, "1", msg.GetErrors()[0].Key)
	assert.Equal(t, "Not found error", msg.GetErrors()[0].Value)
	assert.Equal(t, "3", msg.GetErrors()[1].Key)
	assert.Equal(t, "Bad request error", msg.GetErrors()[1].Value)
}

func TestImport(t *testing.T) {
	ctx := context.Background()
	type Input struct {
		Elements []int `json:"elements"`
	}

	data := []byte(`{"elements": [1, 2, 3]}`)
	checkDatafn := func(input Input) error { return nil }
	dataFn := func(input Input) []int {
		return input.Elements
	}
	actionFn := func(ctx context.Context, input Input, element int) (int, error, error) {
		if element == 2 {
			return 0, errors.New(apperr.Translate(apperr.ErrBadRequest, translator.RU.String())), nil
		}
		if element == 3 {
			return 0, nil, apperr.ErrDBInternal
		}
		return element, nil, nil
	}

	msg, err := task.Import(ctx, data, checkDatafn, dataFn, actionFn)

	assert.NoError(t, err)
	assert.Equal(t, 3, msg.GetCount())
	assert.Len(t, msg.GetErrors(), 2)
	assert.Len(t, msg.GetSuccesses(), 1)
	assert.Equal(t, "1", msg.GetErrors()[0].Key)
	assert.Equal(t, "Bad request error", msg.GetErrors()[0].Value)
}

func TestExport(t *testing.T) {
	ctx := context.Background()
	type Input struct{}
	type Output struct {
		ID int
	}

	checkDatafn := func(input Input) error { return nil }
	listFn := func(ctx context.Context, input Input) ([]int, error) {
		return []int{1, 2, 3}, nil
	}
	actionFn := func(item int) (Output, error) {
		if item == 3 {
			return Output{}, apperr.ErrBadRequest
		}
		return Output{item}, nil
	}

	emptyListError := apperr.ErrBadRequest
	data := []byte(`{}`)

	msg, err := task.Export(ctx, data, emptyListError, checkDatafn, listFn, actionFn)

	assert.NoError(t, err)
	assert.Equal(t, 3, msg.GetCount())
	assert.Len(t, msg.GetErrors(), 1)
	assert.NotNil(t, msg.GetData())
	assert.Equal(t, "2", msg.GetErrors()[0].Key)
	assert.Equal(t, "Bad request error", msg.GetErrors()[0].Value)
}

func TestMassDelete_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	data := []byte(`invalid json`)

	checkDatafn := func(input struct{}) error { return nil }
	idsFn := func(input struct{}) []int { return nil }
	pluckIDsFn := func(ctx context.Context, ids []int) ([]int, error) { return nil, nil }
	actionFn := func(ctx context.Context, input struct{}, id int) error { return nil }

	msg, err := task.MassDelete(ctx, data, errors.New("not found"), checkDatafn, idsFn, pluckIDsFn, actionFn)

	assert.Nil(t, msg)
	assert.Error(t, err)
}

func TestMassDelete_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем контекст

	data := []byte(`{"ids": [1, 2, 3]}`)
	checkDatafn := func(input struct{ IDs []int }) error { return nil }
	idsFn := func(input struct{ IDs []int }) []int { return input.IDs }
	pluckIDsFn := func(ctx context.Context, ids []int) ([]int, error) { return ids, nil }
	actionFn := func(ctx context.Context, input struct{ IDs []int }, id int) error { return nil }

	msg, err := task.MassDelete(ctx, data, errors.New("not found"), checkDatafn, idsFn, pluckIDsFn, actionFn)

	assert.NoError(t, err)
	assert.Equal(t, 3, msg.GetCount())
	assert.Empty(t, msg.GetErrors())
	assert.Empty(t, msg.GetSuccesses())
}

func TestMassUpdate_PluckIDsFnError(t *testing.T) {
	ctx := context.Background()
	data := []byte(`{"ids": [1, 2, 3]}`)
	checkDatafn := func(input struct{ IDs []int }) error { return nil }
	idsFn := func(input struct{ IDs []int }) []int { return input.IDs }
	pluckIDsFn := func(ctx context.Context, ids []int) ([]int, error) {
		return nil, errors.New("pluck error")
	}
	actionFn := func(ctx context.Context, id int, input struct{ IDs []int }) error { return nil }

	msg, err := task.MassUpdate(ctx, data, errors.New("not found"), checkDatafn, idsFn, pluckIDsFn, actionFn)

	assert.Nil(t, msg)
	assert.Error(t, err)
	assert.Equal(t, "pluck error", err.Error())
}

func TestImport_AllSuccess(t *testing.T) {
	ctx := context.Background()
	type Input struct {
		Elements []int `json:"elements"`
	}

	data := []byte(`{"elements": [1, 2, 3]}`)
	checkDatafn := func(input Input) error { return nil }
	dataFn := func(input Input) []int { return input.Elements }
	actionFn := func(ctx context.Context, input Input, element int) (int, error, error) {
		return element, nil, nil
	}

	msg, err := task.Import(ctx, data, checkDatafn, dataFn, actionFn)

	assert.NoError(t, err)
	assert.Equal(t, 3, msg.GetCount())
	assert.Empty(t, msg.GetErrors())
	assert.Len(t, msg.GetSuccesses(), 3)
}

func TestExport_EmptyList(t *testing.T) {
	ctx := context.Background()
	type Input struct {
		Filter string `json:"filter"`
	}

	checkDatafn := func(input Input) error { return nil }
	listFn := func(ctx context.Context, input Input) ([]int, error) {
		return nil, nil
	}
	actionFn := func(item int) (string, error) { return "", nil }

	emptyListError := errors.New("empty list")
	data := []byte(`{"filter": "empty"}`)

	msg, err := task.Export(ctx, data, emptyListError, checkDatafn, listFn, actionFn)

	assert.Nil(t, msg)
	assert.Error(t, err)
	assert.Equal(t, "empty list", err.Error())
}

func TestMassDelete_PluckIDsFnError(t *testing.T) {
	ctx := context.Background()
	data := []byte(`{"ids": [1, 2, 3]}`)
	checkDatafn := func(input struct{ IDs []int }) error { return nil }
	idsFn := func(input struct{ IDs []int }) []int { return input.IDs }
	pluckIDsFn := func(ctx context.Context, ids []int) ([]int, error) {
		return nil, errors.New("pluckIDsFn error")
	}
	actionFn := func(ctx context.Context, input struct{ IDs []int }, id int) error { return nil }

	msg, err := task.MassDelete(ctx, data, errors.New("not found"), checkDatafn, idsFn, pluckIDsFn, actionFn)

	assert.Nil(t, msg)
	assert.Error(t, err)
	assert.Equal(t, "pluckIDsFn error", err.Error())
}

func TestMassUpdate_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	data := []byte(`invalid json`) // Некорректный JSON

	checkDatafn := func(input struct{}) error { return nil }
	idsFn := func(input struct{}) []int { return nil }
	pluckIDsFn := func(ctx context.Context, ids []int) ([]int, error) { return nil, nil }
	actionFn := func(ctx context.Context, id int, input struct{}) error {
		return nil
	}

	msg, err := task.MassUpdate(ctx, data, errors.New("not found"), checkDatafn, idsFn, pluckIDsFn, actionFn)

	assert.Nil(t, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character") // Проверка на сообщение об ошибке JSON
}

func TestMassUpdate_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Отменяем контекст

	data := []byte(`{"ids": [1, 2, 3]}`)
	checkDatafn := func(input struct{}) error { return nil }
	idsFn := func(input struct{}) []int { return nil }
	pluckIDsFn := func(ctx context.Context, ids []int) ([]int, error) { return ids, nil }
	actionFn := func(ctx context.Context, id int, input struct{}) error { return nil }

	msg, err := task.MassUpdate(ctx, data, errors.New("not found"), checkDatafn, idsFn, pluckIDsFn, actionFn)

	assert.NoError(t, err)
	assert.Equal(t, 0, msg.GetCount())
	assert.Empty(t, msg.GetErrors())
	assert.Empty(t, msg.GetSuccesses())
}

func TestExport_ListFnError(t *testing.T) {
	ctx := context.Background()
	type Input struct {
		Filter string `json:"filter"`
	}

	checkDatafn := func(input Input) error { return nil }
	listFn := func(ctx context.Context, input Input) ([]int, error) {
		return nil, errors.New("listFn error")
	}
	actionFn := func(item int) (string, error) { return "", nil }

	emptyListError := errors.New("empty list")
	data := []byte(`{"filter": "test"}`)

	msg, err := task.Export(ctx, data, emptyListError, checkDatafn, listFn, actionFn)

	assert.Nil(t, msg)
	assert.Error(t, err)
	assert.Equal(t, "listFn error", err.Error())
}
