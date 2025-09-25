package resty_logger

import (
	"encoding/json"
	"fmt"
	"strings"

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
				Body:   string(JsonHideImportantData([]byte(req.Body), "pass", "token", "pwd", "code", "secret")),
			},
			Response: LoggerFieldsResponse{
				StatusCode: res.Status,
				Duration:   fmt.Sprintf("%.3fms", float64(res.Duration)/1e6),
				Body:       string(JsonHideImportantData([]byte(res.Body), "pass", "token", "pwd", "code", "secret")),
			},
		}

		jsonBytes, err := json.Marshal(srt)
		if err != nil {
			return ""
		}

		return string(jsonBytes)
	}
}

func JsonHideImportantData(input []byte, keys ...string) []byte {
	if len(keys) == 0 {
		return input
	}

	if input == nil {
		return input
	}

	var data interface{}
	if err := json.Unmarshal(input, &data); err != nil {
		return input
	}

	maskSensitiveFields(data, keys...)

	output, err := json.Marshal(data)
	if err != nil {
		return input
	}

	return output
}

func maskSensitiveFields(data interface{}, keys ...string) {
	switch t := data.(type) {
	case map[string]interface{}:
		for key, value := range t {
			for _, sensitiveKey := range keys {
				if strings.Contains(strings.ToLower(key), strings.ToLower(sensitiveKey)) {
					if _, ok := value.(string); ok {
						t[key] = "****"
						break
					}
				}
			}

			if nested, ok := value.(map[string]interface{}); ok {
				maskSensitiveFields(nested, keys...)
			} else if array, ok := value.([]interface{}); ok {
				for _, item := range array {
					if nestedMap, ok := item.(map[string]interface{}); ok {
						maskSensitiveFields(nestedMap, keys...)
					}
				}
			}
		}

		data = t
	case []interface{}:
		for _, value := range t {
			if nested, ok := value.(map[string]interface{}); ok {
				maskSensitiveFields(nested, keys...)
			} else if array, ok := value.([]interface{}); ok {
				for _, item := range array {
					if nestedMap, ok := item.(map[string]interface{}); ok {
						maskSensitiveFields(nestedMap, keys...)
					}
				}
			}
		}
		data = t
	}
}
