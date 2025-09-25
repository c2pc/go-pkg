package mw

import (
	"net/http"
	"strconv"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"

	"github.com/gin-gonic/gin"
)

func CorsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header(
			"Access-Control-Expose-Headers",
			"Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar,X-Operation-Id,X-Total-Count,X-Broker",
		) // Cross-domain key settings allow browsers to resolve.
		c.Header(
			"Access-Control-Max-Age",
			"172800",
		) // Cache request information in seconds.
		c.Header(
			"Access-Control-Allow-Credentials",
			"false",
		) //  Whether cross-domain requests need to carry cookie information, the default setting is true.
		c.Header(
			"content-type",
			"application/json",
		) // Set the return format to json.
		// Release all option pre-requests
		if c.Request.Method == http.MethodOptions {
			c.JSON(http.StatusOK, "Options Request!")
			c.Abort()
			return
		}
		c.Next()
	}
}

func GinParseOperationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.JSON(http.StatusOK, "Options Request!")
			c.Abort()
			return
		}

		operationID := c.Request.Header.Get(constant.OperationIDHeader)
		if operationID == "" {
			operationID = strconv.FormatInt(time.Now().UnixMilli(), 10)
		}

		ctx := c.Request.Context()
		ctx = mcontext.WithOperationIDContext(ctx, operationID)

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
