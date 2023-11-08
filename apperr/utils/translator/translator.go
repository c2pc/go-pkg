package translator

import (
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/ru"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	rutranslations "github.com/go-playground/validator/v10/translations/ru"
)

var utrans *ut.UniversalTranslator

func SetValidateTranslators(validate *validator.Validate) {
	en := en.New()
	ru := ru.New()
	utrans = ut.New(en, en, ru)
	trans, _ := utrans.GetTranslator("ru")
	_ = rutranslations.RegisterDefaultTranslations(validate, trans)
}

func GetTranslator(acceptLang string) ut.Translator {
	t, found := utrans.FindTranslator(acceptLang)
	if !found {
		t, _ = utrans.FindTranslator("ru")
	}

	return t
}
