package request

import (
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TxHandle(c *gin.Context) *gorm.DB {
	txHandle, exists := c.Get(constant.TxValue)
	if !exists {
		return nil
	}

	return txHandle.(*gorm.DB)
}
