package resty_logger

import (
	"encoding/json"
	"fmt"
	"strings"

	"resty.dev/v3"
)

func limitString(s string, limit int) string {
	if len(s) > limit {
		return s[:limit]
	}
	return s
}

func DebugLogFormatterFunc(dl *resty.DebugLog) string {
	debugLog := "\n"

	req := dl.Request
	debugLog += "~~~ REQUEST ~~~\n" +
		fmt.Sprintf("HOST          : %s  %s%s\n", req.Method, req.Host, req.URI) +
		fmt.Sprintf("OPERATION-ID  : %s\n", req.Header.Get("X-Operation-Id")) +
		fmt.Sprintf("BODY          : %s\n", limitString(string(JsonHideImportantData([]byte(req.Body), "pass", "token", "pwd", "code")), 1000))

	res := dl.Response
	debugLog += "~~~ RESPONSE ~~~\n" +
		fmt.Sprintf("STATUS    : %s\n", res.Status) +
		fmt.Sprintf("DURATION  : %v\n", res.Duration) +
		fmt.Sprintf("BODY      : %v\n", limitString(string(JsonHideImportantData([]byte(res.Body), "pass", "token", "pwd", "code")), 1000))

	return debugLog
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
