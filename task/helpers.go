package task

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/c2pc/go-pkg/v2/task/model"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/datautil"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/jszwec/csvutil"
)

func MassDelete[T any](ctx context.Context, data []byte, notFoundError error, idsFn func(T) []int, pluckIDsFn func(context.Context, []int) ([]int, error), actionFn func(context.Context, T, int) error) (*model.Message, error) {
	msg := model.NewMessage()

	var input T
	err := json.Unmarshal(data, &input)
	if err != nil {
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
		msg.AddError(strconv.Itoa(id), apperr.Translate(notFoundError, translator.EN.String()))
		continue
	}

	for _, id := range ids {
		if ctx.Err() != nil {
			return msg, nil
		}

		err = actionFn(ctx, input, id)
		if err != nil {
			msg.AddError(strconv.Itoa(id), apperr.Translate(err, translator.EN.String()))
			continue
		} else {
			msg.AddSuccess(strconv.Itoa(id))
		}
	}

	return msg, nil
}

func MassUpdate[T any](ctx context.Context, data []byte, notFoundError error, idsFn func(T) []int, pluckIDsFn func(context.Context, []int) ([]int, error), actionFn func(context.Context, int, T) error) (*model.Message, error) {
	msg := model.NewMessage()

	var input T
	err := json.Unmarshal(data, &input)
	if err != nil {
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
		msg.AddError(strconv.Itoa(id), apperr.Translate(notFoundError, translator.EN.String()))
		continue
	}

	for _, id := range ids {
		if ctx.Err() != nil {
			return msg, nil
		}

		err = actionFn(ctx, id, input)
		if err != nil {
			msg.AddError(strconv.Itoa(id), apperr.Translate(err, translator.EN.String()))
			continue
		} else {
			msg.AddSuccess(strconv.Itoa(id))
		}
	}

	return msg, nil
}

func Import[T, C any](ctx context.Context, data []byte, dataFn func(T) []C, actionFn func(context.Context, T, C) (string, error, error)) (*model.Message, error) {
	msg := model.NewMessage()

	var input T
	err := json.Unmarshal(data, &input)
	if err != nil {
		return nil, err
	}

	elements := dataFn(input)

	msg.SetCount(len(elements))

	for i, element := range elements {
		if ctx.Err() != nil {
			return msg, nil
		}

		key, prevErr, err := actionFn(ctx, input, element)
		if err != nil {
			msg.AddError(strconv.Itoa(i), apperr.Translate(err, translator.EN.String()))
			continue
		} else if prevErr != nil {
			msg.AddError(strconv.Itoa(i), prevErr.Error())
			continue
		} else {
			msg.AddSuccess(key)
		}
	}

	return msg, nil
}

func Export[T, C, N any](ctx context.Context, data []byte, emptyListError error, listFn func(context.Context, N) ([]T, error), actionFn func(T) (C, error)) (*model.Message, error) {
	msg := model.NewMessage()

	var input N
	err := json.Unmarshal(data, &input)
	if err != nil {
		return nil, err
	}

	list, err := listFn(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, emptyListError
	}

	msg.SetCount(len(list))

	var export []C
	for i, item := range list {
		if ctx.Err() != nil {
			return msg, nil
		}

		c, err := actionFn(item)
		if err != nil {
			msg.AddError(strconv.Itoa(i), apperr.Translate(err, translator.EN.String()))
		}

		export = append(export, c)
	}

	b, err := csvutil.Marshal(export)
	if err != nil {
		return nil, err
	}

	msg.SetData(b)

	return msg, nil
}
