package middleware

import (
	"database/sql"
	"fmt"
	"github.com/c2pc/go-pkg/apperr"
	"github.com/c2pc/go-pkg/apperr/utils/appErrors"
	"github.com/c2pc/go-pkg/apperr/utils/translate"
	"github.com/c2pc/go-pkg/apperr/x/httperr"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

var (
	ErrInternalMethod = apperr.New("all",
		apperr.WithTitleTranslate(translate.Translate{translate.RU: "Ошибка"}),
		apperr.WithContext("all"),
	)

	ErrCommitDatabaseID = "commit_database_error"
	ErrPanicID          = "panic_error"
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
				httperr.Response(c, ErrInternalMethod.WithError(appErrors.ErrInternal.NewID(ErrPanicID)))
				fmt.Println(r)
				return
			}
		}()

		c.Set("db_trx", txHandle)
		c.Next()

		if statusInList(c.Writer.Status(), []int{http.StatusOK, http.StatusCreated, http.StatusNoContent}) {
			if err := txHandle.Commit().Error; err != nil {
				httperr.Response(c, ErrInternalMethod.WithError(appErrors.ErrInternal.NewID(ErrCommitDatabaseID)))
				return
			}
		} else {
			txHandle.Rollback()
		}
	}
}
