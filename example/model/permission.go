package model

import (
	"github.com/c2pc/go-pkg/v2/example/i18n"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
)

var Permissions = []model2.Permission{
	{Method: "books", Desc: i18n.BooksPermission},
	{Method: "books/likes", Desc: i18n.BooksLikesPermission},
	{Method: "news", Desc: i18n.NewsPermission},
}
