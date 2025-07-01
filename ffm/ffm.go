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
	Downloader
	SetAddr(addr string) FileManager
	LS(ctx context.Context, request LSRequest) ([]FileInfo, error)
	Info(ctx context.Context, path string) (*FileInfo, error)
	MkDir(ctx context.Context, request MkDirRequest) (*FileInfo, error)
	DecodeAudio(ctx context.Context, request DecodeAudioRequest) (*FileInfo, error)
	Upload(ctx context.Context, request UploadRequest) ([]FileInfo, error)
	Remove(ctx context.Context, request RemoveRequest) error
	GenDownloadPath(info FileInfo) string
	GenCompressDownloadPath(info FileInfo) string
	CP(ctx context.Context, input FileCopyRequest) (*FileInfo, error)
	MV(ctx context.Context, input FileMoveRequest) (*FileInfo, error)
	Unpack(ctx context.Context, request FileUnpackRequest) (*FileInfo, error)
}

func (f *FFM) Unpack(ctx context.Context, request FileUnpackRequest) (*FileInfo, error) {
	var response FileInfo
	err := f.jsonRequest(ctx, http.MethodPost, "unpack", request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

type FFM struct {
	addr    string
	service string
}

type Config struct {
	Addr    string
	Service string
}

func New(cfg Config) (FileManager, error) {
	if cfg.Service == "" {
		return nil, errors.New("empty file manager service")
	}

	if len(cfg.Addr) > 0 && cfg.Addr[len(cfg.Addr)-1:] != "/" {
		cfg.Addr += "/"
	}

	ffm := &FFM{
		addr:    cfg.Addr,
		service: cfg.Service,
	}

	return ffm, nil
}

func (f *FFM) CP(ctx context.Context, input FileCopyRequest) (*FileInfo, error) {
	var response FileInfo
	err := f.jsonRequest(ctx, http.MethodPost, "cp", input, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (f *FFM) MV(ctx context.Context, input FileMoveRequest) (*FileInfo, error) {
	var response FileInfo
	err := f.jsonRequest(ctx, http.MethodPost, "mv", input, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (f *FFM) SetAddr(addr string) FileManager {
	if len(addr) > 0 && addr[len(addr)-1:] != "/" {
		addr += "/"
	}

	f.addr = addr
	return f
}

func (f *FFM) LS(ctx context.Context, request LSRequest) ([]FileInfo, error) {
	var response []FileInfo
	err := f.jsonRequest(ctx, http.MethodGet, "ls", request, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (f *FFM) Info(ctx context.Context, path string) (*FileInfo, error) {
	request := PathRequest{
		Path: path,
	}

	var response FileInfo
	err := f.jsonRequest(ctx, http.MethodGet, "info", request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *FFM) MkDir(ctx context.Context, request MkDirRequest) (*FileInfo, error) {
	var response FileInfo
	err := f.jsonRequest(ctx, http.MethodPost, "mkdir", request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *FFM) DecodeAudio(ctx context.Context, request DecodeAudioRequest) (*FileInfo, error) {
	var response FileInfo
	err := f.jsonRequest(ctx, http.MethodPost, "decode-audio", request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *FFM) Upload(ctx context.Context, request UploadRequest) ([]FileInfo, error) {
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

	if logger.IsDebugEnabled(level.TEST) {
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

	var output []FileInfo
	if status, err := parseResult(resp, &output); err != nil {
		if logger.IsDebugEnabled(level.TEST) {
			logger.Infof("RESPONSE - %s - %s - %s - %+v - %d", operationID, method, url, err, status)
		}

		return nil, err
	}

	if logger.IsDebugEnabled(level.TEST) {
		logger.Infof("RESPONSE - %s - %s - %s - %+v", operationID, method, url, output)
	}

	return output, nil
}

func (f *FFM) GenDownloadPath(info FileInfo) string {
	return f.addr + "api/v1/" + f.service + "/download?path=" + info.Path
}

func (f *FFM) GenCompressDownloadPath(info FileInfo) string {
	return f.addr + "api/v1/" + f.service + "/compress-download?path=" + info.Path
}

func (f *FFM) Remove(ctx context.Context, request RemoveRequest) error {
	err := f.jsonRequest(ctx, http.MethodPost, "remove", request, nil)
	if err != nil {
		return err
	}

	return nil
}
