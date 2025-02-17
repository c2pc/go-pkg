package ffm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/google/go-querystring/query"
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

func (f *FFM) request(ctx context.Context, method, url string, operationID string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, f.addr+"api/v1/"+f.service+"/"+url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Operation-Id", operationID)

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (f *FFM) jsonRequest(ctx context.Context, method, url string, input interface{}, output interface{}) error {
	var reqBody []byte
	var err error
	if input != nil {
		if method == http.MethodGet {
			v, err := query.Values(input)
			if err != nil {
				return err
			}

			url = url + "?" + v.Encode()
		} else {
			reqBody, err = json.Marshal(input)
			if err != nil {
				return err
			}
		}
	}

	operationID := strconv.Itoa(int(time.Now().UnixMilli()))
	op := ctx.Value(constant.OperationID)
	if op2, ok := op.(string); ok {
		operationID = op2
	}

	if level.Is(f.debug, level.TEST) {
		logger.Infof("REQUEST - %s - %s - %s - %+v", operationID, method, url, string(reqBody))
	}

	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := f.request(ctx2, method, url, operationID, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if status, err := parseResult(resp, output); err != nil {
		if level.Is(f.debug, level.TEST) {
			logger.Infof("RESPONSE - %s - %s - %s - %+v - %d", operationID, method, url, err, status)
		}

		return err
	}

	if level.Is(f.debug, level.TEST) {
		logger.Infof("RESPONSE - %s - %s - %s - %+v", operationID, method, url, output)
	}

	return nil
}
