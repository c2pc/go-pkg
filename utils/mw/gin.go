package mw

import (
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/response/httperr"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrEmptyOperationID = apperr.New("empty_operation_id_header",
		apperr.WithTextTranslate(translator.Translate{
			translator.RU: "Неверный запрос",
			translator.EN: "Invalid request",
		}),
		apperr.WithCode(code.InvalidArgument),
	)
)

// CorsHandler gin cross-domain configuration.
func CorsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header(
			"Access-Control-Expose-Headers",
			"Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar",
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

		if c.Request.Method == http.MethodPost {
			operationID := c.Request.Header.Get(constant.OperationID)
			if operationID == "" {
				httperr.Response(c, ErrEmptyOperationID)
				c.Abort()
				return
			}

			ctx := c.Request.Context()
			ctx = mcontext.WithOperationIDContext(ctx, operationID)

			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}
