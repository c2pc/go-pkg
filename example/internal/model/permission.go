package model

import (
	"github.com/c2pc/go-pkg/v2/example/internal/i18n"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
)

var Permissions = []model2.Permission{
	{Method: "news", Desc: i18n.NewsPermission},
}
