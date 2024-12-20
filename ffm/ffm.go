package ffm

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/c2pc/go-pkg/v2/utils/logger"
)

type FileManager interface {
	LS(ctx context.Context, request LSRequest) ([]FileInfoResponse, error)
	Info(ctx context.Context, path string) (*FileInfoResponse, error)
	MkDir(ctx context.Context, request MkDirRequest) (*FileInfoResponse, error)
	DecodeMp3(ctx context.Context, request DecodeMp3Request) (*FileInfoResponse, error)
	Upload(ctx context.Context, request UploadRequest) ([]FileInfoResponse, error)
	Remove(ctx context.Context, request RemoveRequest) error
}

type FFM struct {
	addr    string
	service string
	debug   string
}

func New(addr, service, debug string) (FileManager, error) {
	if addr == "" {
		return nil, errors.New("empty file manager url")
	}
	if service == "" {
		return nil, errors.New("empty file manager service")
	}

	if addr[len(addr)-1:] != "/" {
		addr += "/"
	}

	ffm := &FFM{
		addr:    addr,
		service: service,
		debug:   debug,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := ffm.LS(ctx, LSRequest{})
	if err != nil {
		return ffm, err
	}

	return ffm, nil
}

func (f *FFM) LS(ctx context.Context, request LSRequest) ([]FileInfoResponse, error) {
	var response []FileInfoResponse
	err := f.jsonRequest(ctx, http.MethodGet, "ls", request, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (f *FFM) Info(ctx context.Context, path string) (*FileInfoResponse, error) {
	request := PathRequest{
		Path: path,
	}

	var response FileInfoResponse
	err := f.jsonRequest(ctx, http.MethodGet, "info", request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *FFM) MkDir(ctx context.Context, request MkDirRequest) (*FileInfoResponse, error) {
	var response FileInfoResponse
	err := f.jsonRequest(ctx, http.MethodPost, "mkdir", request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *FFM) DecodeMp3(ctx context.Context, request DecodeMp3Request) (*FileInfoResponse, error) {
	var response FileInfoResponse
	err := f.jsonRequest(ctx, http.MethodPost, "decode-mp3", request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *FFM) Upload(ctx context.Context, request UploadRequest) ([]FileInfoResponse, error) {
	method := http.MethodPost
	url := "upload"

	operationID := strconv.Itoa(int(time.Now().UnixMilli()))
	op := ctx.Value(constant.OperationID)
	if op2, ok := op.(string); ok {
		operationID = op2
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	if err := w.WriteField("path", request.Path); err != nil {
		return nil, err
	}

	if err := w.WriteField("append", strconv.FormatBool(request.Append)); err != nil {
		return nil, err
	}

	for _, file := range request.Files {
		fw, err := w.CreateFormFile("files", filepath.Base(file.Name))
		if err != nil {
			return nil, err
		}

		if _, err = io.Copy(fw, file.Reader); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	if level.Is(f.debug, level.TEST) {
		r := map[string]interface{}{"path": request.Path, "files": len(request.Files), "append": request.Append}
		logger.Infof("REQUEST - %s - %s - %s - files -> %v", operationID, method, url, r)
	}

	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := f.request(ctx2, method, url, operationID, w.FormDataContentType(), &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var output []FileInfoResponse
	if _, err = parseResult(resp, &output); err != nil {
		if level.Is(f.debug, level.TEST) {
			logger.Infof("RESPONSE - %s - %s - %s - %+v", operationID, method, url, err)
		}

		return nil, err
	}

	if level.Is(f.debug, level.TEST) {
		logger.Infof("RESPONSE - %s - %s - %s - %+v", operationID, method, url, output)
	}

	return output, nil
}

func (f *FFM) Remove(ctx context.Context, request RemoveRequest) error {
	err := f.jsonRequest(ctx, http.MethodPost, "remove", request, nil)
	if err != nil {
		return err
	}

	return nil
}
