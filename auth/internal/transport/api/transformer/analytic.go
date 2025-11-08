package transformer

import (
	"bytes"
	"compress/gzip"
	"io"
	"time"

	"github.com/aws/smithy-go/ptr"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
	"github.com/gin-gonic/gin"
)

func decompressGzip(compressedData []byte) string {
	if compressedData == nil {
		return ""
	}
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return ""
	}
	defer reader.Close()

	uncompressedData, err := io.ReadAll(reader)
	if err != nil {
		return ""
	}
	return string(uncompressedData)
}

type AnalyticTransformer struct {
	ID           int64     `json:"id"`
	UserID       *int64    `json:"user_id"`
	Login        string    `json:"login"`
	Name         string    `json:"name"`
	Action       string    `json:"action"`
	Error        string    `json:"error"`
	ClientIP     string    `json:"client_ip"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	StatusCode   int       `json:"status_code"`
	OperationID  string    `json:"operation_id"`
	CreatedAt    time.Time `json:"created_at"`
	RequestBody  string    `json:"request_body"`
	ResponseBody string    `json:"response_body"`
}

func AnalyticTransform(m *model.Analytic) AnalyticTransformer {
	return AnalyticTransformer{
		ID:           m.ID,
		OperationID:  m.OperationID,
		Path:         m.Path,
		Method:       m.Method,
		StatusCode:   m.StatusCode,
		ClientIP:     m.ClientIP,
		RequestBody:  decompressGzip(m.RequestBody),
		ResponseBody: decompressGzip(m.ResponseBody),
		CreatedAt:    m.CreatedAt,
		UserID:       m.UserID,
		Error:        ptr.ToString(m.Error),
		Action:       ptr.ToString(m.Action),
		Login:        ptr.ToString(m.Login),
		Name:         ptr.ToString(m.Name),
	}
}

type AnalyticListTransformer struct {
	ID          int64     `json:"id"`
	UserID      *int64    `json:"user_id"`
	Login       string    `json:"login"`
	Name        string    `json:"name"`
	Action      string    `json:"action"`
	Error       string    `json:"error"`
	ClientIP    string    `json:"client_ip"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	StatusCode  int       `json:"status_code"`
	OperationID string    `json:"operation_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func AnalyticListTransform(c *gin.Context, p *meta.Pagination[model.Analytic]) []AnalyticListTransformer {
	transformer.PaginationTransform(c, p)

	r := make([]AnalyticListTransformer, 0)

	for _, m := range p.Rows {
		r = append(r, AnalyticListTransformer{
			ID:          m.ID,
			Path:        m.Path,
			OperationID: m.OperationID,
			Method:      m.Method,
			StatusCode:  m.StatusCode,
			ClientIP:    m.ClientIP,
			CreatedAt:   m.CreatedAt,
			UserID:      m.UserID,
			Error:       ptr.ToString(m.Error),
			Action:      ptr.ToString(m.Action),
			Login:       ptr.ToString(m.Login),
			Name:        ptr.ToString(m.Name),
		})
	}

	return r
}
