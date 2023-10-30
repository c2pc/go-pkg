package apperr

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/ru"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	rutranslations "github.com/go-playground/validator/v10/translations/ru"
	"golang.org/x/text/language"
)

var utrans *ut.UniversalTranslator

func SetTranslators(validate *validator.Validate) {
	en := en.New()
	ru := ru.New()
	utrans = ut.New(en, en, ru)
	trans, _ := utrans.GetTranslator("ru")
	_ = rutranslations.RegisterDefaultTranslations(validate, trans)
}

func getTranslator(acceptLang string) ut.Translator {
	t, found := utrans.FindTranslator(acceptLang)
	if !found {
		t, _ = utrans.FindTranslator("ru")
	}

	return t
}

func getTranslatorHTTP(c *gin.Context) ut.Translator {
	acceptLang := c.GetHeader("Accept-Language")
	var matcher = language.NewMatcher([]language.Tag{
		language.Russian,
		language.MustParse("ru-RU"),
		language.English,
		language.MustParse("en-US"),
	})
	tag, _ := language.MatchStrings(matcher, acceptLang)
	matched, _, _ := matcher.Match(tag)
	base, _ := matched.Base()
	return getTranslator(base.String())
}

func getTranslateHTTP(c *gin.Context) string {
	acceptLang := c.GetHeader("Accept-Language")
	var matcher = language.NewMatcher([]language.Tag{
		language.Russian,
		language.MustParse("ru-RU"),
		language.English,
		language.MustParse("en-US"),
	})
	tag, _ := language.MatchStrings(matcher, acceptLang)
	matched, _, _ := matcher.Match(tag)
	base, _ := matched.Base()
	return base.String()
}

type Translate map[string]string

func (t Translate) Translate(acceptLang string, args ...interface{}) string {
	tr, found := t[acceptLang]
	if !found {
		tr, found = t["ru"]
		if !found {
			return ""
		}
	}

	return fmt.Sprintf(tr, args...)
}

func (t Translate) TranslateHttp(c *gin.Context, args ...interface{}) string {
	acceptLang := c.GetHeader("Accept-Language")
	var matcher = language.NewMatcher([]language.Tag{
		language.Russian,
		language.MustParse("ru-RU"),
		language.English,
		language.MustParse("en-US"),
	})
	tag, _ := language.MatchStrings(matcher, acceptLang)
	matched, _, _ := matcher.Match(tag)
	base, _ := matched.Base()
	return t.Translate(base.String(), args...)
}
