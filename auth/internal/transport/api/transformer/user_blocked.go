package transformer

import (
	"time"

	"github.com/c2pc/go-pkg/v2/auth/internal/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/transformer"
	"github.com/gin-gonic/gin"
)

type UserBlockedListTransformer struct {
	ID         int64     `json:"id"`
	Login      string    `json:"login"`
	FirstName  string    `json:"first_name"`
	SecondName *string   `json:"second_name"`
	LastName   *string   `json:"last_name"`
	IsDomain   bool      `json:"is_domain"`
	BlockedAt  time.Time `json:"blocked_at"`
	UnLockedAt time.Time `json:"unlocked_at"`
}

func UserBlockedListTransform(c *gin.Context, limiter *fx.LimiterHolder, p *meta.Pagination[model.UserBlocked]) []UserBlockedListTransformer {
	transformer.PaginationTransform(c, p)

	r := make([]UserBlockedListTransformer, 0)

	for _, m := range p.Rows {
		r = append(r, UserBlockedListTransformer{
			ID:         m.User.ID,
			Login:      m.User.Login,
			FirstName:  m.User.FirstName,
			SecondName: m.User.SecondName,
			LastName:   m.User.LastName,
			IsDomain:   m.User.IsDomain,
			BlockedAt:  m.BlockedAt,
			UnLockedAt: m.BlockedAt.Add(limiter.Get().BlockingTTL),
		})
	}

	return r
}
