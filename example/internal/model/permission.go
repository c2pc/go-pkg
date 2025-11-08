package model

import (
	"github.com/c2pc/go-pkg/v2/example/internal/i18n"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/translator"
)

var Permissions = []meta.Permission{
	{Method: "news", Desc: i18n.NewsPermission},
	{Method: "stream", Desc: i18n.StreamPermission},
	{Method: "query-history", Desc: translator.Translate{translator.RU: "История запросов", translator.EN: "Query history"}},
}
