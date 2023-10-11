package middleware

import (
	"database/sql"
	"github.com/c2pc/go-pkg/apperr"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type ITransaction interface {
	DBTransactionMiddleware() gin.HandlerFunc
}

type Transaction struct {
	DB *gorm.DB
}

func NewTr(db *gorm.DB) *Transaction {
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

func (tr *Transaction) DBTransactionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		txHandle := tr.DB.
			WithContext(c.Request.Context()).
			Begin(&sql.TxOptions{})

		defer func() {
			if r := recover(); r != nil {
				txHandle.Rollback()
				apperr.HTTPResponse(c, apperr.ErrInternal)
				return
			}
		}()

		c.Set("db_trx", txHandle)
		c.Next()

		if statusInList(c.Writer.Status(), []int{http.StatusOK, http.StatusCreated, http.StatusNoContent}) {
			if err := txHandle.Commit().Error; err != nil {
				apperr.HTTPResponse(c, apperr.ErrInternal.WithError(err))
				return
			}
		} else {
			txHandle.Rollback()
		}
	}
}
