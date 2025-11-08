package resty_logger

import (
	"encoding/json"
	"fmt"

	"github.com/c2pc/go-pkg/v2/utils/jsonutil"
	"resty.dev/v3"
)

type LoggerFields struct {
	Request  LoggerFieldsRequest  `json:"request"`
	Response LoggerFieldsResponse `json:"response"`
}

type LoggerFieldsRequest struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	Body   string `json:"body"`
}
type LoggerFieldsResponse struct {
	StatusCode string `json:"status_code"`
	Duration   string `json:"duration"`
	Body       string `json:"body"`
}

func DebugLogFormatterFunc() func(dl *resty.DebugLog) string {
	return func(dl *resty.DebugLog) string {
		req := dl.Request
		res := dl.Response

		srt := LoggerFields{
			Request: LoggerFieldsRequest{
				Method: req.Method,
				URL:    fmt.Sprintf("%s%s", req.Host, req.URI),
				Body:   string(jsonutil.JsonHideImportantData([]byte(req.Body), "pass", "token", "pwd", "code", "secret")),
			},
			Response: LoggerFieldsResponse{
				StatusCode: res.Status,
				Duration:   fmt.Sprintf("%.3fms", float64(res.Duration)/1e6),
				Body:       string(jsonutil.JsonHideImportantData([]byte(res.Body), "pass", "token", "pwd", "code", "secret")),
			},
		}

		jsonBytes, err := json.Marshal(srt)
		if err != nil {
			return ""
		}

		return string(jsonBytes)
	}
}
