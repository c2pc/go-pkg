package model

import (
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/translator"
)

var Permissions = []model2.Permission{
	{Method: "books", Desc: translator.Translate{translator.RU: "Книги", translator.EN: "Books"}},
	{Method: "books/likes", Desc: translator.Translate{translator.RU: "Книги/Лайки", translator.EN: "Books/Likes"}},
}
