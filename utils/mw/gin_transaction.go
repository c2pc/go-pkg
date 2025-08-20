package mw

import (
	"database/sql"
	"net/http"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
			logger.ErrorfLog(c.Request.Context(), constant.APP_ID, "%s - %v", apperr.ErrInternal.Error(), r)
			response.Response(c, apperr.ErrInternal)
			return
		}
	}()

	c.Set(string(constant.TxValue), txHandle)
	c.Next()

	if statusInList(c.Writer.Status(), []int{http.StatusOK, http.StatusCreated, http.StatusNoContent, http.StatusFound}) {
		if err := txHandle.Commit().Error; err != nil {
			response.Response(c, apperr.ErrDBInternal.WithError(err))
			return
		}
	} else {
		txHandle.Rollback()
	}
}
