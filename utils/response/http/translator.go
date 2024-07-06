package http

import (
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"golang.org/x/text/language"
)

func GetTranslate(c *gin.Context) string {
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

func getTranslator(c *gin.Context) ut.Translator {
	return translator.GetTranslator(GetTranslate(c))
}
