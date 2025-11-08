package model

type Version struct {
	AppName string
	App     string `json:"app"`
	DB      string `json:"db"`
}

type Migration struct {
	Version string `json:"version"`
	Dirty   bool   `json:"dirty"`
}

func (m Migration) TableName() string {
	return "schema_migrations"
}

func (m Migration) TableNameAuth() string {
	return "schema_auth_migrations"
}
