package ffm

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/resty_logger"
	tls2 "github.com/c2pc/go-pkg/v2/utils/tls"
	"resty.dev/v3"
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
	err := f.request(ctx, http.MethodPost, "unpack", request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

type FFM struct {
	service string
	client  *resty.Client
}

type Config struct {
	Addr    string
	Service string
}

func New(cfg Config) (FileManager, error) {
	if cfg.Service == "" {
		return nil, errors.New("empty file manager service")
	}

	tlsConfig := &tls.Config{
		CipherSuites:       tls2.GetCipherSuiteIDs(),
		InsecureSkipVerify: true,
	}

	client := resty.New().
		SetBaseURL(strings.TrimSuffix(cfg.Addr, "/") + "/api/v1/").
		SetTLSClientConfig(tlsConfig).
		SetTimeout(10 * time.Second).
		SetDebug(true).
		SetDebugLogFormatter(resty_logger.DebugLogFormatterFunc()).
		SetLogger(&resty_logger.RestyLogger{LoggerID: "FFM"})

	ffm := &FFM{
		client:  client,
		service: cfg.Service,
	}

	return ffm, nil
}

func (f *FFM) CP(ctx context.Context, input FileCopyRequest) (*FileInfo, error) {
	var response FileInfo
	err := f.request(ctx, http.MethodPost, "cp", input, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (f *FFM) MV(ctx context.Context, input FileMoveRequest) (*FileInfo, error) {
	var response FileInfo
	err := f.request(ctx, http.MethodPost, "mv", input, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (f *FFM) SetAddr(addr string) FileManager {
	f.client.SetBaseURL(strings.TrimSuffix(addr, "/") + "/api/v1/")
	return f
}

func (f *FFM) LS(ctx context.Context, request LSRequest) ([]FileInfo, error) {
	var response []FileInfo
	err := f.request(ctx, http.MethodGet, "ls", request, &response)
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
	err := f.request(ctx, http.MethodGet, "info", request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *FFM) MkDir(ctx context.Context, request MkDirRequest) (*FileInfo, error) {
	var response FileInfo
	err := f.request(ctx, http.MethodPost, "mkdir", request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *FFM) DecodeAudio(ctx context.Context, request DecodeAudioRequest) (*FileInfo, error) {
	var response FileInfo
	err := f.request(ctx, http.MethodPost, "decode-audio", request, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (f *FFM) Upload(ctx context.Context, request UploadRequest) ([]FileInfo, error) {
	method := http.MethodPost
	url := "upload"

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

	opID := strconv.Itoa(int(time.Now().UnixMicro()))
	if op, ok := mcontext.GetOperationID(ctx); ok {
		parts := strings.Split(op, "-")
		if len(parts) > 0 {
			opID = parts[len(parts)-1]
		}
	}

	var output []FileInfo
	req := f.client.R().
		SetContext(ctx).
		SetContentType("application/json").
		SetHeader("X-Operation-Id", opID).
		SetResult(output)

	resp, err := req.Execute(method, url)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "connection refused") {
			return output, apperr.ErrServerIsNotAvailable.WithError(err)
		}
		return output, err
	}

	if resp.IsSuccess() {
		return output, nil
	}

	switch resp.StatusCode() {
	case 400:
		return output, apperr.ErrValidation
	case 401:
		return output, apperr.ErrUnauthenticated
	case 403:
		return output, apperr.ErrForbidden
	case 404:
		return output, apperr.ErrNotFound
	default:
		return output, apperr.ErrInternal
	}
}

func (f *FFM) GenDownloadPath(info FileInfo) string {
	return f.client.BaseURL() + f.service + "/download?path=" + url.PathEscape(info.Path)
}

func (f *FFM) GenCompressDownloadPath(info FileInfo) string {
	return f.client.BaseURL() + f.service + "/compress-download?path=" + url.PathEscape(info.Path) + "&type=zip"
}

func (f *FFM) Remove(ctx context.Context, request RemoveRequest) error {
	err := f.request(ctx, http.MethodPost, "remove", request, nil)
	if err != nil {
		return err
	}

	return nil
}
