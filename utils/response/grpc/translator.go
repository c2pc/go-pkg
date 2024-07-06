package grpc

import (
	"context"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	ut "github.com/go-playground/universal-translator"
	"golang.org/x/text/language"
)

func GetTranslate(ctx context.Context) string {
	acceptLang, ok := ctx.Value("accept-language").(string)
	if !ok {
		acceptLang = "ru"
	}

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

func getTranslator(ctx context.Context) ut.Translator {
	return translator.GetTranslator(GetTranslate(ctx))
}
