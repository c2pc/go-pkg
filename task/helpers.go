package task

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	model2 "github.com/c2pc/go-pkg/v2/task/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/secret"

	"github.com/c2pc/go-pkg/v2/task/model"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/datautil"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/jszwec/csvutil"
)

type CheckDataFn[T any] func(T) error
type IdsFn[T any, C string | int] func(T) []C
type PluckIDsFn[C string | int] func(context.Context, []C) ([]C, error)
type DeleteActionFn[T any, C string | int] func(context.Context, T, C) error

type UpdateActionFn[T any, C string | int] func(context.Context, C, T) error

type DataFn[T, C any] func(T) []C
type ImportActionFn[T, C any, D string | int] func(context.Context, T, C) (D, error, error)

type ListFn[T, N any] func(context.Context, N) ([]T, error)
type ExportActionFn[T, C any] func(T) (C, error)

func MassDelete[T any, C string | int](ctx context.Context, taskID int, msgChan chan<- *model.Message, data []byte, notFoundError error, checkDataFn CheckDataFn[T], idsFn IdsFn[T, C], pluckIDsFn PluckIDsFn[C], actionFn DeleteActionFn[T, C]) (*model.Message, error) {
	msg := model.NewMessage()

	var input T
	err := json.Unmarshal(data, &input)
	if err != nil {
		return nil, apperr.ErrInternal.WithError(err)
	}

	err = checkDataFn(input)
	if err != nil {
		if ctx.Err() != nil {
			return msg, nil
		}
		return nil, err
	}

	ids := idsFn(input)

	msg.SetCount(len(ids))

	pluckedIds, err := pluckIDsFn(ctx, ids)
	if err != nil {
		if ctx.Err() != nil {
			return msg, nil
		}
		return nil, err
	}

	singleIds := datautil.Single[C](ids, pluckedIds)
	for _, id := range singleIds {
		msg.AddError(idToString(id), apperr.Translate(notFoundError, translator.RU.String()))
		continue
	}

	for i, id := range pluckedIds {
		if ctx.Err() != nil {
			return msg, nil
		}

		var ctx2 = ctx
		if opID, ok := mcontext.GetOperationID(ctx); ok {
			ctx2 = mcontext.WithOperationIDContext(ctx, opID+"-"+idToString(id))
		}

		err = actionFn(ctx2, input, id)
		if err != nil {
			msg.AddError(idToString(id), apperr.Translate(err, translator.RU.String()))
		} else {
			msg.AddSuccess(idToString(id))
		}

		if i%100 == 0 {
			msgChan <- msg
		}
	}

	return msg, nil
}

func MassUpdate[T any, C string | int](ctx context.Context, taskID int, msgChan chan<- *model.Message, data []byte, notFoundError error, checkDataFn CheckDataFn[T], idsFn IdsFn[T, C], pluckIDsFn PluckIDsFn[C], actionFn UpdateActionFn[T, C]) (*model.Message, error) {
	msg := model.NewMessage()

	var input T
	err := json.Unmarshal(data, &input)
	if err != nil {
		return nil, apperr.ErrInternal.WithError(err)
	}

	err = checkDataFn(input)
	if err != nil {
		if ctx.Err() != nil {
			return msg, nil
		}
		return nil, err
	}

	ids := idsFn(input)

	msg.SetCount(len(ids))

	pluckedIds, err := pluckIDsFn(ctx, ids)
	if err != nil {
		if ctx.Err() != nil {
			return msg, nil
		}
		return nil, err
	}

	singleIds := datautil.Single(ids, pluckedIds)
	for _, id := range singleIds {
		msg.AddError(idToString(id), apperr.Translate(notFoundError, translator.RU.String()))
		continue
	}

	for i, id := range pluckedIds {
		if ctx.Err() != nil {
			return msg, nil
		}

		var ctx2 = ctx
		if opID, ok := mcontext.GetOperationID(ctx); ok {
			ctx2 = mcontext.WithOperationIDContext(ctx, opID+"-"+idToString(id))
		}

		err = actionFn(ctx2, id, input)
		if err != nil {
			msg.AddError(idToString(id), apperr.Translate(err, translator.RU.String()))
		} else {
			msg.AddSuccess(idToString(id))
		}

		if i%100 == 0 {
			msgChan <- msg
		}
	}

	return msg, nil
}

func Import[T, C any, D string | int](ctx context.Context, taskID int, msgChan chan<- *model.Message, data []byte, checkDataFn CheckDataFn[T], dataFn DataFn[T, C], actionFn ImportActionFn[T, C, D]) (*model.Message, error) {
	msg := model.NewMessage()

	var input T
	err := json.Unmarshal(data, &input)
	if err != nil {
		return nil, apperr.ErrInternal.WithError(err)
	}

	err = checkDataFn(input)
	if err != nil {
		if ctx.Err() != nil {
			return msg, nil
		}
		return nil, err
	}

	elements := dataFn(input)

	msg.SetCount(len(elements))

	for i, element := range elements {
		if ctx.Err() != nil {
			return msg, nil
		}

		var ctx2 = ctx
		if opID, ok := mcontext.GetOperationID(ctx); ok {
			ctx2 = mcontext.WithOperationIDContext(ctx, opID+"-"+idToString(i))
		}

		key, prevErr, err := actionFn(ctx2, input, element)

		k := strconv.Itoa(i)
		if idToString(key) != "" {
			k = idToString(key)
		}

		if err != nil {
			msg.AddError(k, apperr.Translate(err, translator.RU.String()))
		} else if prevErr != nil {
			msg.AddError(k, prevErr.Error())
		} else {
			msg.AddSuccess(k)
		}

		if i%100 == 0 {
			msgChan <- msg
		}
	}

	return msg, nil
}

func Export[T, C, N any](ctx context.Context, taskID int, msgChan chan<- *model.Message, data []byte, emptyListError error, checkDataFn CheckDataFn[N], listFn ListFn[T, N], actionFn ExportActionFn[T, C]) (*model.Message, error) {
	msg := model.NewMessage()

	var input N
	err := json.Unmarshal(data, &input)
	if err != nil {
		return nil, apperr.ErrInternal.WithError(err)
	}

	err = checkDataFn(input)
	if err != nil {
		if ctx.Err() != nil {
			return msg, nil
		}
		return nil, err
	}

	list, err := listFn(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, emptyListError
	}

	rand, _ := secret.GenerateRandomString(5)
	fileName := fmt.Sprintf("export_%d%s.csv", taskID, rand)
	file, err := CreateFile(fileName)
	if err != nil {
		return nil, apperr.ErrInternal.WithError(err)
	}
	defer file.Close()

	msg.FileName = &fileName

	msgChan <- msg

	w := csv.NewWriter(file)
	defer w.Flush()
	enc := csvutil.NewEncoder(w)

	for i, item := range list {
		msg.SetCount(i + 1)

		if ctx.Err() != nil {
			return msg, nil
		}

		c, err := actionFn(item)
		if err != nil {
			fmt.Println(err)
			msg.AddError(strconv.Itoa(i), apperr.Translate(err, translator.RU.String()))
			continue
		}
		fmt.Println(c)

		if err := enc.Encode(c); err != nil {
			return msg, apperr.ErrInternal.WithError(err)
		}

		if i%100 == 0 {
			msgChan <- msg
		}
	}

	return msg, nil
}

func idToString[C string | int](id C) string {
	switch v := any(id).(type) {
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	default:
		return "unsupported type"
	}
}

func CreateFile(fileName string) (*os.File, error) {
	path := model2.MediaPath + "/" + fileName

	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	return file, nil
}
