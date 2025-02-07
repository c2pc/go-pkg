package ffm

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
)

var (
	ErrServerIsNotUnavailable    = apperr.New("ffm_server_is_not_unavailable", apperr.WithTextTranslate(translator.Translate{translator.RU: "Сервер FFM недоступен", translator.EN: "Server FFM is unavailable"}), apperr.WithCode(code.Unavailable))
	ErrGenerateLink              = apperr.New("err_to_generate_link", apperr.WithTextTranslate(translator.Translate{translator.RU: "Ошибка генерации ссылки", translator.EN: "Error to generate link"}), apperr.WithCode(code.InvalidArgument))
	ErrCheckLink                 = apperr.New("err_to_check_link", apperr.WithTextTranslate(translator.Translate{translator.RU: "Ошибка проверки ссылки", translator.EN: "Error to check link"}), apperr.WithCode(code.InvalidArgument))
	ErrServiceNotFound           = apperr.New("service_not_found", apperr.WithTextTranslate(translator.Translate{translator.RU: "Сервис не найден", translator.EN: "Service not found"}), apperr.WithCode(code.NotFound))
	ErrInvalidFilter             = apperr.New("service_invalid_filter", apperr.WithTextTranslate(translator.Translate{translator.RU: "Неправильный фильтр", translator.EN: "Invalid filter"}), apperr.WithCode(code.InvalidArgument))
	ErrObjectIsNotFile           = apperr.New("service_object_is_not_file", apperr.WithTextTranslate(translator.Translate{translator.RU: "Объект не является файлом", translator.EN: "Object is not file"}), apperr.WithCode(code.InvalidArgument))
	ErrInvalidFileExt            = apperr.New("service_invalid_file_ext", apperr.WithTextTranslate(translator.Translate{translator.RU: "Неверное расширение файла", translator.EN: "Invalid file extension"}), apperr.WithCode(code.InvalidArgument))
	ErrObjectIsNotDir            = apperr.New("service_object_is_not_dir", apperr.WithTextTranslate(translator.Translate{translator.RU: "Объект не является директорией", translator.EN: "Object is not dir"}), apperr.WithCode(code.InvalidArgument))
	ErrObjectIsNotFound          = apperr.New("service_object_is_not_found", apperr.WithTextTranslate(translator.Translate{translator.RU: "Объект не найден", translator.EN: "Object is not found"}), apperr.WithCode(code.InvalidArgument))
	ErrCreateDirectory           = apperr.New("service_create_directory", apperr.WithTextTranslate(translator.Translate{translator.RU: "Невозможно создать директорию", translator.EN: "Can't create directory"}), apperr.WithCode(code.InvalidArgument))
	ErrObjectRemove              = apperr.New("service_object_remove", apperr.WithTextTranslate(translator.Translate{translator.RU: "Невозможно удалить объект", translator.EN: "Can't remove object"}), apperr.WithCode(code.InvalidArgument))
	ErrObjectIsAlreadyExists     = apperr.New("service_object_is_already_exists", apperr.WithTextTranslate(translator.Translate{translator.RU: "Объект уже добавлен", translator.EN: "Object is already exists"}), apperr.WithCode(code.InvalidArgument))
	ErrObjectNoWritePermissions  = apperr.New("service_object_no_write_permissions", apperr.WithTextTranslate(translator.Translate{translator.RU: "Нет прав на запись", translator.EN: "No write permissions"}), apperr.WithCode(code.InvalidArgument))
	ErrObjectNoRemovePermissions = apperr.New("service_object_no_remove_permissions", apperr.WithTextTranslate(translator.Translate{translator.RU: "Нет прав на удаление", translator.EN: "No remove permissions"}), apperr.WithCode(code.InvalidArgument))
	ErrDecode                    = apperr.New("service_decode", apperr.WithTextTranslate(translator.Translate{translator.RU: "Ошибка декодирования", translator.EN: "Decode error"}), apperr.WithCode(code.InvalidArgument))
	ErrRemoveObject              = apperr.New("service_remove_object", apperr.WithTextTranslate(translator.Translate{translator.RU: "Ошибка удаления объекта", translator.EN: "Remove object error"}), apperr.WithCode(code.InvalidArgument))
	ErrNoPermissionsToRemoveFile = apperr.New("service_no_permissions_to_remove_file", apperr.WithTextTranslate(translator.Translate{translator.RU: "Нет прав на удаление файла", translator.EN: "No permissions to remove file"}), apperr.WithCode(code.InvalidArgument))
	ErrNoPermissionsToCreateFile = apperr.New("service_no_permissions_to_create_file", apperr.WithTextTranslate(translator.Translate{translator.RU: "Нет прав на создание файла", translator.EN: "No permissions to create file"}), apperr.WithCode(code.InvalidArgument))
	ErrNoPermissionsToRemoveDir  = apperr.New("service_no_permissions_to_remove_dir", apperr.WithTextTranslate(translator.Translate{translator.RU: "Нет прав на удаление директории", translator.EN: "No permissions to remove dir"}), apperr.WithCode(code.InvalidArgument))
	ErrNoPermissionsToCreateDir  = apperr.New("service_no_permissions_to_create_dir", apperr.WithTextTranslate(translator.Translate{translator.RU: "Нет прав на создание директории", translator.EN: "No permissions to create dir"}), apperr.WithCode(code.InvalidArgument))
	ErrNoPermissionsToDownload   = apperr.New("service_no_permissions_to_download", apperr.WithTextTranslate(translator.Translate{translator.RU: "Нет прав на загрузку", translator.EN: "No permissions to download"}), apperr.WithCode(code.InvalidArgument))
)

type ErrorResponse struct {
	ID     string        `json:"id"`
	Text   string        `json:"text"`
	Errors []interface{} `json:"errors"`
}

func parseResult(resp *http.Response, output interface{}) (int, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if output != nil && respBody != nil {
			err = json.Unmarshal(respBody, &output)
			if err != nil {
				return 0, err
			}
		}

		return resp.StatusCode, nil
	}

	var errorResponse ErrorResponse
	_ = json.Unmarshal(respBody, &errorResponse)

	err = apperr.ErrInternal
	if errorResponse.ID != "" {
		if errorResponse.Errors != nil {
			v, _ := json.Marshal(errorResponse.Errors)
			err = apperr.New(errorResponse.ID, apperr.WithText(errorResponse.Text)).WithErrorText(string(v))
		} else {
			err = apperr.New(errorResponse.ID, apperr.WithText(errorResponse.Text))
		}
	}

	var err2 apperr.Error
	switch true {
	case apperr.Is(err, ErrServiceNotFound):
		err2 = ErrServiceNotFound
	case apperr.Is(err, ErrInvalidFilter):
		err2 = ErrInvalidFilter
	case apperr.Is(err, ErrObjectIsNotFile):
		err2 = ErrObjectIsNotFile
	case apperr.Is(err, ErrInvalidFileExt):
		err2 = ErrInvalidFileExt
	case apperr.Is(err, ErrObjectIsNotDir):
		err2 = ErrObjectIsNotDir
	case apperr.Is(err, ErrObjectIsNotFound):
		err2 = ErrObjectIsNotFound
	case apperr.Is(err, ErrCreateDirectory):
		err2 = ErrCreateDirectory
	case apperr.Is(err, ErrObjectIsAlreadyExists):
		err2 = ErrObjectIsAlreadyExists
	case apperr.Is(err, ErrObjectRemove):
		err2 = ErrObjectRemove
	case apperr.Is(err, ErrObjectNoWritePermissions):
		err2 = ErrObjectNoWritePermissions
	case apperr.Is(err, ErrObjectNoRemovePermissions):
		err2 = ErrObjectNoRemovePermissions
	case apperr.Is(err, ErrDecode):
		err2 = ErrDecode
	case apperr.Is(err, ErrRemoveObject):
		err2 = ErrRemoveObject
	case apperr.Is(err, ErrNoPermissionsToRemoveFile):
		err2 = ErrNoPermissionsToRemoveFile
	case apperr.Is(err, ErrNoPermissionsToCreateFile):
		err2 = ErrNoPermissionsToCreateFile
	case apperr.Is(err, ErrNoPermissionsToRemoveDir):
		err2 = ErrNoPermissionsToRemoveDir
	case apperr.Is(err, ErrNoPermissionsToCreateDir):
		err2 = ErrNoPermissionsToCreateDir
	case apperr.Is(err, ErrNoPermissionsToDownload):
		err2 = ErrNoPermissionsToDownload
	case apperr.Is(err, apperr.ErrBadRequest):
		err2 = apperr.ErrBadRequest
	case apperr.Is(err, apperr.ErrValidation):
		err2 = apperr.ErrBadRequest
	default:
		if resp.StatusCode >= 500 {
			err2 = ErrServerIsNotUnavailable
		} else {
			err2 = apperr.ErrInternal
		}
	}

	return resp.StatusCode, err2.WithError(err)
}
