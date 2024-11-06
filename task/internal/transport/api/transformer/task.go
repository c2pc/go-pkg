package transformer

import (
	"encoding/json"

	"github.com/c2pc/go-pkg/v2/task/internal/model"
	model3 "github.com/c2pc/go-pkg/v2/task/model"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
	"github.com/gin-gonic/gin"
)

type SimpleTaskTransformer struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func SimpleTaskTransform(m *model.Task) *SimpleTaskTransformer {
	return &SimpleTaskTransformer{
		ID:        m.ID,
		Name:      m.Name,
		Status:    m.Status,
		Type:      m.Type,
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

type TaskTransformer struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	UserID    int             `json:"user_id"`
	Status    string          `json:"status"`
	Type      string          `json:"type"`
	Message   *model3.Message `json:"message,omitempty"`
	FileSize  *int64          `json:"file_size,omitempty"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

func TaskTransform(m *model.Task) *TaskTransformer {
	var msg model3.Message
	if m.Output != nil {
		_ = json.Unmarshal(m.Output, &msg)
	}
	r := &TaskTransformer{
		ID:        m.ID,
		Name:      m.Name,
		UserID:    m.UserID,
		Status:    m.Status,
		Type:      m.Type,
		Message:   &msg,
		FileSize:  m.FileSize,
		CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return r
}

type TaskListTransformer struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func TaskListTransform(c *gin.Context, p *model2.Pagination[model.Task]) []TaskListTransformer {
	transformer.PaginationTransform(c, p)

	r := make([]TaskListTransformer, 0)

	for _, m := range p.Rows {
		t := TaskListTransformer{
			ID:        m.ID,
			Name:      m.Name,
			Status:    m.Status,
			Type:      m.Type,
			CreatedAt: m.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: m.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		r = append(r, t)
	}

	return r
}
