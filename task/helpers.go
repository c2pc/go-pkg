package task

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"

	model2 "github.com/c2pc/go-pkg/v2/task/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/logger"
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
		msg.SetCount(i + 1)
		if ctx.Err() != nil {
			return msg, nil
		}

		var ctx2 = ctx
		if opID, ok := mcontext.GetOperationID(ctx); ok {
			ctx2 = mcontext.WithOperationIDContext(ctx, opID+"-"+idToString(id))
		}

		func() {
			defer func() {
				if err := recover(); err != nil {
					logger.WarningfLog(ctx2, "TASK", "%v", err)
					msg.AddError(idToString(id), apperr.Translate(apperr.ErrInternal, translator.RU.String()))
				}
			}()

			err = actionFn(ctx2, input, id)
			if err != nil {
				msg.AddError(idToString(id), apperr.Translate(err, translator.RU.String()))
			} else {
				msg.AddSuccess(idToString(id))
			}
		}()

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
		msg.SetCount(i + 1)
		if ctx.Err() != nil {
			return msg, nil
		}

		var ctx2 = ctx
		if opID, ok := mcontext.GetOperationID(ctx); ok {
			ctx2 = mcontext.WithOperationIDContext(ctx, opID+"-"+idToString(id))
		}

		func() {
			defer func() {
				if err := recover(); err != nil {
					logger.WarningfLog(ctx2, "TASK", "%v", err)
					msg.AddError(idToString(id), apperr.Translate(apperr.ErrInternal, translator.RU.String()))
				}
			}()

			err = actionFn(ctx2, id, input)
			if err != nil {
				msg.AddError(idToString(id), apperr.Translate(err, translator.RU.String()))
			} else {
				msg.AddSuccess(idToString(id))
			}
		}()

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

	for i, element := range elements {
		msg.SetCount(i + 1)
		if ctx.Err() != nil {
			return msg, nil
		}

		var ctx2 = ctx
		if opID, ok := mcontext.GetOperationID(ctx); ok {
			ctx2 = mcontext.WithOperationIDContext(ctx, opID+"-"+idToString(i))
		}

		k := strconv.Itoa(i)

		func() {
			defer func() {
				if err := recover(); err != nil {
					logger.WarningfLog(ctx2, "TASK", "%v", err)
					msg.AddError(k, apperr.Translate(apperr.ErrInternal, translator.RU.String()))
				}
			}()

			key, prevErr, err := actionFn(ctx2, input, element)

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
		}()

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
	file, _, err := CreateFile(fileName)
	if err != nil {
		return nil, apperr.ErrInternal.WithError(err)
	}
	defer file.Close()

	msg.FileName = &fileName

	msgChan <- msg

	w := csv.NewWriter(file)
	defer w.Flush()
	enc := csvutil.NewEncoder(w)

	var export []C
	for i, item := range list {
		msg.SetCount(i + 1)

		if ctx.Err() != nil {
			return msg, nil
		}

		func() {
			defer func() {
				if err := recover(); err != nil {
					logger.WarningfLog(ctx, "TASK", "%v", err)
					msg.AddError(strconv.Itoa(i), apperr.Translate(apperr.ErrInternal, translator.RU.String()))
				}
			}()

			c, err := actionFn(item)
			if err != nil {
				msg.AddError(strconv.Itoa(i), apperr.Translate(err, translator.RU.String()))
			}

			export = append(export, c)
		}()

		if i%100 == 0 {
			if err := enc.Encode(export); err != nil {
				return msg, apperr.ErrInternal.WithError(err)
			}
			export = []C{}
			msgChan <- msg
		}
	}

	if len(export) > 0 {
		if err := enc.Encode(export); err != nil {
			return msg, apperr.ErrInternal.WithError(err)
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

func CreateFile(fileName string) (*os.File, string, error) {
	filePath := path.Join(model2.MediaPath, fileName)

	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return nil, "", err
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, "", err
	}

	return file, filePath, nil
}
