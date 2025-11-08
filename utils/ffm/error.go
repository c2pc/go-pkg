package ffm

import (
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
)

type ErrorResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

var (
	ErrServerIsNotUnavailable    = apperr.New("ffm_server_is_not_unavailable", apperr.WithTextTranslate(translator.Translate{translator.RU: "Сервис flat-file недоступен", translator.EN: "Server flat-file is unavailable"}), apperr.WithCode(code.Unavailable))
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
