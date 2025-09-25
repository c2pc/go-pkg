package model

import (
	"github.com/c2pc/go-pkg/v2/example/internal/i18n"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/translator"
)

var Permissions = []model2.Permission{
	{Method: "news", Desc: i18n.NewsPermission},
	{Method: "stream", Desc: i18n.StreamPermission},
	{Method: "query-history", Desc: translator.Translate{translator.RU: "История запросов", translator.EN: "Query history"}},
}
