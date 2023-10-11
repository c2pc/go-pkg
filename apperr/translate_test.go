package apperr

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetTranslators(t *testing.T) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		SetTranslators(v)
	}

	if utrans == nil {
		t.Errorf("utrans is empty")
	}

	_, found := utrans.FindTranslator("ru")
	if !found {
		t.Errorf("ru translator not found")
	}

	_, found = utrans.FindTranslator("en")
	if !found {
		t.Errorf("en translator not found")
	}
}

func TestGetTranslators(t *testing.T) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		SetTranslators(v)
	}
	if utrans == nil {
		t.Errorf("utrans is empty")
	}

	gin.SetMode(gin.TestMode)
	for _, al := range []string{"ru", "en", "fa"} {
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		if c == nil {
			t.Errorf("error to create gin")
			return
		}

		c.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
		c.Request.Header.Set("Accept-Language", al)

		tr := getTranslatorHTTP(c)
		tr2, found := utrans.FindTranslator(al)
		if !found {
			tr2, found = utrans.FindTranslator("ru")
			if !found {
				t.Errorf("%s translator not found", al)
			}
		}

		if tr != tr2 {
			t.Errorf("%s translator not found", al)
		}
	}
}

func TestTranslate(t *testing.T) {
	tr := Translate{
		"ru": "RU",
		"en": "EN",
	}

	gin.SetMode(gin.TestMode)
	for _, al := range []string{"ru", "en", "fa"} {
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		if c == nil {
			t.Errorf("error to create gin")
			return
		}

		c.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
		c.Request.Header.Set("Accept-Language", al)

		text := tr.TranslateHttp(c)
		tx, found := tr[al]
		if !found {
			tx, _ = tr["ru"]
		}

		if text != tx {
			t.Errorf("translate not found")
		}
	}
}
