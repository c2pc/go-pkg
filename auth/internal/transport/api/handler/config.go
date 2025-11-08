package handler

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/crewjam/saml/samlsp"

	"github.com/gin-gonic/gin"
)

const maxSize = 1 << 20 // 1 MB = 1 * 1024 * 1024 байт

var (
	ErrKeyRequired = apperr.New("key_is_required",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Ключ не найден", translator.EN: "Key not found"}),
		apperr.WithCode(code.InvalidArgument),
	)
	ErrParsingFile = apperr.New("parsing_file_error",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Ошибка анализа файла %s: %s", translator.EN: "Error parsing %s file: %s"}),
		apperr.WithCode(code.InvalidArgument),
	)
)

type ConfigHandler struct {
	authConfigService service.IConfigService
	tr                mw.ITransaction
}

func NewConfigHandlers(
	authConfigService service.IConfigService,
	tr mw.ITransaction,
) *ConfigHandler {
	return &ConfigHandler{
		authConfigService,
		tr,
	}
}

func (h *ConfigHandler) GetService() service.IConfigService {
	return h.authConfigService
}

func (h *ConfigHandler) Init(secured *gin.RouterGroup) {
	authConfig := secured.Group("configs")
	{
		authConfig.PATCH("/:key", h.tr.DBTransaction, h.Update)
		authConfig.GET("/:key", h.GetByKey)
		authConfig.POST("/saml.metadata", h.tr.DBTransaction, h.uploadSamlMetadata)
		authConfig.POST("/saml.cert", h.tr.DBTransaction, h.uploadSamlCert)
	}
}

func (h *ConfigHandler) GetByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.Response(c, ErrKeyRequired)
		return
	}

	data, err := h.authConfigService.GetByKey(c.Request.Context(), key)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.ConfigTransform(data))
}

func (h *ConfigHandler) Update(c *gin.Context) {
	c.Request = mcontext.WithOpActionRequest(c.Request, "Изменение конфигурации системы")
	key := c.Param("key")
	if key == "" {
		response.Response(c, ErrKeyRequired)
		return
	}

	cred, err := request2.BindJSON[json.RawMessage](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	if cred == nil {
		response.Response(c, apperr.ErrEmptyData)
		return
	}

	action, err := h.authConfigService.Trx(request2.TxHandle(c)).Update(c.Request.Context(), key, *cred)
	if err != nil {
		c.Request = mcontext.WithOpActionRequest(c.Request, action)
		response.Response(c, err)
		return
	}
	c.Request = mcontext.WithOpActionRequest(c.Request, action)

	c.Status(http.StatusOK)
}

func (h *ConfigHandler) uploadSamlMetadata(c *gin.Context) {
	c.Request = mcontext.WithOpActionRequest(c.Request, "Изменение конфигурации системы")

	file, err := c.FormFile("metadata")
	if err != nil {
		response.Response(c, apperr.ErrBadRequest.WithError(err))
		return
	}

	if file.Size > maxSize {
		response.Response(c, apperr.ErrBadRequest.WithErrorText("metadata file size too large"))
		return
	}

	dat, err := file.Open()
	if err != nil {
		response.Response(c, apperr.ErrBadRequest.WithError(err))
		return
	}
	defer dat.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(dat); err != nil {
		response.Response(c, apperr.ErrBadRequest.WithError(err))
		return
	}

	_, err = samlsp.ParseMetadata(buf.Bytes())
	if err != nil {
		response.Response(c, ErrParsingFile.WithTextArgs(file.Filename, err.Error()))
		return
	}

	_, action, err := h.authConfigService.Trx(request2.TxHandle(c)).UploadFiles(c.Request.Context(), model.AuthConfigKey, []service.UploadFilesInput{
		{
			Key:      model.SAMLMetadataConfigFileKey,
			FileName: file.Filename,
			Data:     buf.Bytes(),
		},
	})
	if err != nil {
		c.Request = mcontext.WithOpActionRequest(c.Request, action)
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metadata": file.Filename,
	})
}

func (h *ConfigHandler) uploadSamlCert(c *gin.Context) {
	c.Request = mcontext.WithOpActionRequest(c.Request, "Изменение конфигурации системы")

	cert, err := c.FormFile("cert")
	if err != nil {
		response.Response(c, apperr.ErrBadRequest.WithError(err))
		return
	}

	if cert.Size > maxSize {
		response.Response(c, apperr.ErrBadRequest.WithErrorText("cert file size too large"))
		return
	}

	key, err := c.FormFile("key")
	if err != nil {
		response.Response(c, apperr.ErrBadRequest.WithError(err))
		return
	}

	if key.Size > maxSize {
		response.Response(c, apperr.ErrBadRequest.WithErrorText("key file size too large"))
		return
	}

	certFile, err := cert.Open()
	if err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}
	defer certFile.Close()

	keyFile, err := key.Open()
	if err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}
	defer keyFile.Close()

	var certBuf, keyBuf bytes.Buffer
	if _, err := certBuf.ReadFrom(certFile); err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}

	if _, err := keyBuf.ReadFrom(keyFile); err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		return
	}

	keyPair, err := tls.X509KeyPair(certBuf.Bytes(), keyBuf.Bytes())
	if err != nil {
		response.Response(c, ErrParsingFile.WithTextArgs(cert.Filename, err.Error()))
		return
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		response.Response(c, ErrParsingFile.WithTextArgs(key.Filename, err.Error()))
		return
	}

	_, action, err := h.authConfigService.Trx(request2.TxHandle(c)).UploadFiles(c.Request.Context(), model.AuthConfigKey, []service.UploadFilesInput{
		{
			Key:      model.SAMLCertConfigFileKey,
			FileName: cert.Filename,
			Data:     certBuf.Bytes(),
		}, {
			Key:      model.SAMLKeyConfigFileKey,
			FileName: key.Filename,
			Data:     keyBuf.Bytes(),
		},
	})
	if err != nil {
		c.Request = mcontext.WithOpActionRequest(c.Request, action)
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cert": cert.Filename,
		"key":  key.Filename,
	})
}
