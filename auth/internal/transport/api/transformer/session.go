package transformer

import (
	"time"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
	"github.com/gin-gonic/gin"
)

type SessionListTransformer struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	DeviceID  *string   `json:"device_name"`
	LoggedAt  time.Time `json:"logged_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *UserSimpleTransformer `json:"user"`
}

func SessionListTransform(c *gin.Context, p *meta.Pagination[model.RefreshToken]) []SessionListTransformer {
	transformer.PaginationTransform(c, p)

	r := make([]SessionListTransformer, 0)

	for _, m := range p.Rows {
		var deviceID *string
		if d, ok := model.DeviceID2Name[m.DeviceID]; ok {
			deviceID = &d
		}

		user := SessionListTransformer{
			ID:        m.ID,
			UserID:    m.UserID,
			DeviceID:  deviceID,
			LoggedAt:  m.LoggedAt,
			UpdatedAt: m.UpdatedAt,
			User:      transformer.Nillable(m.User, UserSimpleTransform),
		}

		r = append(r, user)
	}

	return r
}
