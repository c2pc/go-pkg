package transformer

import (
	"time"

	model2 "github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
	"github.com/gin-gonic/gin"
)

type SessionListTransformer struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	DeviceID  *string   `json:"device_name"`
	LoggedAt  time.Time `json:"logged_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *UserSimpleTransformer `json:"user"`
}

func SessionListTransform(c *gin.Context, p *model.Pagination[model2.RefreshToken]) []SessionListTransformer {
	transformer.PaginationTransform(c, p)

	r := make([]SessionListTransformer, 0)

	for _, m := range p.Rows {
		var deviceID *string
		if d, ok := model2.DeviceID2Name[m.DeviceID]; ok {
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
