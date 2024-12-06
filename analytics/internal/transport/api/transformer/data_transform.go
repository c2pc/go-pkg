package transformer

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"time"

	"github.com/c2pc/go-pkg/v2/analytics/internal/models"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
	"github.com/gin-gonic/gin"
)

func decompressGzip(data string) string {
	if data == "" {
		return data
	}
	compressedData := []byte(data)
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return data
	}
	defer reader.Close()

	uncompressedData, err := ioutil.ReadAll(reader)
	if err != nil {
		return data
	}
	return string(uncompressedData)
}

type AnalyticsTransformer struct {
	ID           uint      `json:"id" gorm:"primary_key"`
	OperationID  string    `json:"operation_id"`
	Path         string    `json:"path"`
	Method       string    `json:"method"`
	StatusCode   int       `json:"status_code"`
	ClientIP     string    `json:"client_ip"`
	RequestBody  string    `json:"request_body"`
	ResponseBody string    `json:"response_body"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       int       `json:"user_id"`
	Login        string    `json:"login"`
	FirstName    string    `json:"first_name"`
	SecondName   string    `json:"second_name"`
	LastName     string    `json:"last_name"`
}

func AnalyticTransform(m *models.Analytics) AnalyticsTransformer {
	return AnalyticsTransformer{
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
		Login:        m.User.Login,
		FirstName:    m.User.FirstName,
		SecondName:   m.User.SecondName,
		LastName:     m.User.LastName,
	}
}

type AnalyticsSummaryTransformer struct {
	ID          uint      `json:"id"`
	Path        string    `json:"path"`
	OperationID string    `json:"operation_id"`
	Method      string    `json:"method"`
	StatusCode  int       `json:"status_code"`
	ClientIP    string    `json:"client_ip"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      int       `json:"user_id"`
	Login       string    `json:"login"`
	FirstName   string    `json:"first_name"`
	SecondName  string    `json:"second_name"`
	LastName    string    `json:"last_name"`
}

func AnalyticSummaryTransform(m *models.Analytics) AnalyticsSummaryTransformer {
	return AnalyticsSummaryTransformer{
		ID:          m.ID,
		Path:        m.Path,
		OperationID: m.OperationID,
		Method:      m.Method,
		StatusCode:  m.StatusCode,
		ClientIP:    m.ClientIP,
		CreatedAt:   m.CreatedAt,
		UserID:      m.UserID,
		Login:       m.User.Login,
		FirstName:   m.User.FirstName,
		SecondName:  m.User.SecondName,
		LastName:    m.User.LastName,
	}
}

func AnalyticListTransform(c *gin.Context, p *model2.Pagination[models.Analytics]) []AnalyticsSummaryTransformer {
	transformer.PaginationTransform(c, p)

	r := make([]AnalyticsSummaryTransformer, 0)

	for _, m := range p.Rows {
		analytics := AnalyticSummaryTransform(&m)
		r = append(r, analytics)
	}

	return r
}
