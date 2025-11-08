package ffm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/google/go-querystring/query"
	"resty.dev/v3"
)

type FileUnpackRequest struct {
	Src           string `json:"src"`
	Dst           string `json:"dst"`
	SkipParentDir bool   `json:"skip_parent_dir"`
}

type LSRequest struct {
	Path          string  `json:"path" url:"path"`
	Recursive     bool    `json:"recursive" url:"recursive"`
	FileFilter    *string `json:"file_filter" url:"file_filter"`
	DirFilter     *string `json:"dir_filter" url:"dir_filter"`
	SkipDirs      bool    `json:"skip_dirs" url:"skip_dirs"`
	SkipEmptyDirs bool    `json:"skip_empty_dirs" url:"skip_empty_dirs"`
	SkipPath      bool    `json:"skip_path" url:"skip_path"`
}

type FileCopyRequest struct {
	Src             string `json:"src"`
	Dst             string `json:"dst"`
	ReplaceIfExists bool   `json:"replace_if_exists"`
}

type FileMoveRequest struct {
	Src             string `json:"src"`
	Dst             string `json:"dst"`
	ReplaceIfExists bool   `json:"replace_if_exists"`
}
type PathRequest struct {
	Path string `json:"path" url:"path"`
}

type MkDirRequest struct {
	Path              string `json:"path" url:"path"`
	Dir               string `json:"dir" url:"dir"`
	Recursive         bool   `json:"recursive" url:"recursive"`
	IgnoreExistsError bool   `json:"ignore_exists_error" url:"ignore_exists_error"`
}

type DecodeAudioRequest struct {
	Path            string `json:"path" url:"path"`
	DeleteOriginal  bool   `json:"delete_original" url:"delete_original"`
	ReplaceIfExists bool   `json:"replace_if_exists" url:"replace_if_exists"`
}

type UploadFileRequest struct {
	Name   string
	Reader io.Reader
}

type UploadRequest struct {
	Path   string
	Files  []UploadFileRequest
	Append bool
}

type RemoveRequest struct {
	Path                string `json:"path"`
	IgnoreNotFoundError bool   `json:"ignore_not_found_error"`
	OnlyChildren        bool   `json:"only_children"`
}

func (f *FFM) request(ctx context.Context, method, url string, input interface{}, output interface{}) error {
	opID := strconv.Itoa(int(time.Now().UnixMicro()))
	if op, ok := mcontext.GetOperationID(ctx); ok {
		parts := strings.Split(op, "-")
		if len(parts) > 0 {
			opID = parts[len(parts)-1]
		}
	}

	req := f.client.R().
		SetContext(ctx).
		SetContentType("application/json").
		SetHeader("X-Operation-Id", opID)

	if output != nil {
		req = req.SetResult(output)
	}

	if method == resty.MethodGet && input != nil {
		values, err := query.Values(input)
		if err != nil {
			return err
		}
		req.SetQueryParamsFromValues(values)
	} else if input != nil {
		req.SetBody(input)
	}

	resp, err := req.Execute(method, url)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "connection refused") {
			return apperr.ErrServerIsNotAvailable.WithError(err)
		}
		return err
	}

	if resp.IsSuccess() {
		return nil
	}

	var errorResponse ErrorResponse
	contentType := resp.Header().Get("Content-Type")
	if strings.Contains(contentType, "application/json") && len(resp.Bytes()) > 0 {
		_ = json.Unmarshal(resp.Bytes(), &errorResponse)
	}

	err = apperr.ErrInternal
	if errorResponse.ID != "" {
		err = apperr.New(errorResponse.ID, apperr.WithText(errorResponse.Text))
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
		err2 = apperr.ErrBadRequest.WithError(err)
	case apperr.Is(err, apperr.ErrValidation):
		err2 = apperr.ErrBadRequest.WithError(err)
	default:
		if resp.StatusCode() >= 500 {
			err2 = ErrServerIsNotUnavailable.WithError(err)
		} else {
			err2 = apperr.ErrInternal.WithError(err)
		}
	}

	return err2
}
