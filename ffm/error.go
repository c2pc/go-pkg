package ffm

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
)

var (
	ErrServiceNotFound           = apperr.New("service_not_found")
	ErrInvalidFilter             = apperr.New("service_invalid_filter")
	ErrObjectIsNotFile           = apperr.New("service_object_is_not_file")
	ErrInvalidFileExt            = apperr.New("service_invalid_file_ext")
	ErrObjectIsNotDir            = apperr.New("service_object_is_not_dir")
	ErrObjectIsNotFound          = apperr.New("service_object_is_not_found")
	ErrCreateDirectory           = apperr.New("service_create_directory")
	ErrObjectRemove              = apperr.New("service_object_remove")
	ErrObjectIsAlreadyExists     = apperr.New("service_object_is_already_exists")
	ErrObjectNoWritePermissions  = apperr.New("service_object_no_write_permissions")
	ErrObjectNoRemovePermissions = apperr.New("service_object_no_remove_permissions")
	ErrDecode                    = apperr.New("service_decode")
	ErrRemoveObject              = apperr.New("service_remove_object")
	ErrNoPermissionsToRemoveFile = apperr.New("service_no_permissions_to_remove_file")
	ErrNoPermissionsToCreateFile = apperr.New("service_no_permissions_to_create_file")
	ErrNoPermissionsToRemoveDir  = apperr.New("service_no_permissions_to_remove_dir")
	ErrNoPermissionsToCreateDir  = apperr.New("service_no_permissions_to_create_dir")
	ErrNoPermissionsToDownload   = apperr.New("service_no_permissions_to_download")
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
		err2 = apperr.ErrInternal
	}

	return resp.StatusCode, err2.WithError(err)
}
