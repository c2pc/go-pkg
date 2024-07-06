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
	enLang := en.New()
	ruLang := ru.New()

	utrans = ut.New(enLang, enLang, ruLang)
	trans, _ := utrans.GetTranslator("ru")
	_ = rutranslations.RegisterDefaultTranslations(validate, trans)
}

func GetTranslator(acceptLang string) ut.Translator {
	t, found := utrans.GetTranslator(acceptLang)
	if !found {
		t, _ = utrans.GetTranslator("ru")
	}

	return t
}

func RegisterValidatorTranslation(acceptLang Language, tag, text string, override bool) (string, ut.Translator, validator.RegisterTranslationsFunc, validator.TranslationFunc) {
	return tag,
		GetTranslator(string(acceptLang)),
		func(ut ut.Translator) error {
			return ut.Add(tag, text, override)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T(fe.Tag(), fe.Field())
			if err != nil {
				return fe.Error()
			}
			return t
		}
}
