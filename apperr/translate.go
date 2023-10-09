package apperr

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/ru"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	rutranslations "github.com/go-playground/validator/v10/translations/ru"
)

var utrans *ut.UniversalTranslator

func SetTranslators(validate *validator.Validate) {
	en := en.New()
	ru := ru.New()
	utrans = ut.New(en, en, ru)
	trans, _ := utrans.GetTranslator("ru")
	_ = rutranslations.RegisterDefaultTranslations(validate, trans)
}

func getTranslator(c *gin.Context) ut.Translator {
	acceptLang := c.GetHeader("Accept-Language")
	t, found := utrans.FindTranslator(acceptLang)
	if !found {
		t, _ = utrans.FindTranslator("ru")
	}

	return t
}

type Translate map[string]string

func (t Translate) Translate(c *gin.Context) string {
	acceptLang := c.GetHeader("Accept-Language")
	tr, found := t[acceptLang]
	if !found {
		return t["ru"]
	}

	return tr
}
