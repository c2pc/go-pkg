package transaction

import (
	"database/sql"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/response/httperr"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type ITransaction interface {
	DBTransaction(c *gin.Context)
}

type Transaction struct {
	DB *gorm.DB
}

func NewTransaction(db *gorm.DB) *Transaction {
	return &Transaction{
		DB: db,
	}
}

func statusInList(status int, statusList []int) bool {
	for _, i := range statusList {
		if i == status {
			return true
		}
	}
	return false
}

func (tr *Transaction) DBTransaction(c *gin.Context) {
	txHandle := tr.DB.
		Session(&gorm.Session{NewDB: true}).
		WithContext(c.Request.Context()).
		Begin(&sql.TxOptions{})

	defer func() {
		if r := recover(); r != nil {
			txHandle.Rollback()
			logger.ErrorfLog(c.Request.Context(), constant.DB_ID, "%s - %v", apperr.ErrDBInternal.Error(), r)
			httperr.Response(c, apperr.ErrDBInternal)
			return
		}
	}()

	c.Set(constant.TxValue, txHandle)
	c.Next()

	if statusInList(c.Writer.Status(), []int{http.StatusOK, http.StatusCreated, http.StatusNoContent}) {
		if err := txHandle.Commit().Error; err != nil {
			httperr.Response(c, apperr.ErrDBInternal.WithError(err))
			return
		}
	} else {
		txHandle.Rollback()
	}
}
