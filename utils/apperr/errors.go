package apperr

import (
	"errors"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/i18n"
	"sync"
)

type ErrMap struct {
	mu     sync.RWMutex
	errMap map[string]Error
}

func (e *ErrMap) Get(id string) (Error, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if errFromMap, ok := e.errMap[id]; ok {
		return errFromMap, nil
	}

	return Error{}, errors.New("error not found")
}

func (e *ErrMap) Set(id string, err Error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.errMap[id] = err
}

var (
	ErrMapManager = ErrMap{
		errMap: make(map[string]Error),
		mu:     sync.RWMutex{},
	}
)

// Объявление стандартных ошибок приложения
var (
	ErrSyntax               = New("syntax_error", WithTextTranslate(i18n.ErrSyntax), WithCode(code.InvalidArgument))
	ErrValidation           = New("validation_error", WithTextTranslate(i18n.ErrValidation), WithCode(code.InvalidArgument))
	ErrValidationImport     = New("validation_import_error", WithTextTranslate(i18n.ErrValidation), WithCode(code.InvalidArgument))
	ErrEmptyData            = New("empty_data_error", WithTextTranslate(i18n.ErrEmptyData), WithCode(code.InvalidArgument))
	ErrInternal             = New("internal_error", WithTextTranslate(i18n.ErrInternal), WithCode(code.Internal))
	ErrBadRequest           = New("bad_request_error", WithTextTranslate(i18n.ErrBadRequest), WithCode(code.InvalidArgument))
	ErrForbidden            = New("forbidden_error", WithTextTranslate(i18n.ErrForbidden), WithCode(code.PermissionDenied))
	ErrUnauthenticated      = New("unauthenticated_error", WithTextTranslate(i18n.ErrUnauthenticated), WithCode(code.Unauthenticated))
	ErrNotFound             = New("not_found_error", WithTextTranslate(i18n.ErrNotFound), WithCode(code.NotFound))
	ErrServerIsNotAvailable = New("server_is_not_available", WithTextTranslate(i18n.ErrServerIsNotAvailable), WithCode(code.Unavailable))
	ErrContextCanceled      = New("context_canceled", WithTextTranslate(i18n.ErrContextCanceled), WithCode(code.Unavailable))
)

// Объявление стандартных ошибок базы данных
var (
	ErrDBRecordNotFound = New("db_not_found", WithTextTranslate(i18n.ErrDBRecordNotFound), WithCode(code.NotFound))
	ErrDBDuplicated     = New("db_duplicated", WithTextTranslate(i18n.ErrDBDuplicated), WithCode(code.AlreadyExists))
	ErrDBInternal       = New("db_internal", WithTextTranslate(i18n.ErrDBInternal), WithCode(code.Internal))
)

type ImportError struct {
	Errors map[string]string
}

func (e ImportError) Error() string {
	return "import error"
}
