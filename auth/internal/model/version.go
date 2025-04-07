package model

type Version struct {
	App string `json:"app"`
	DB  string `json:"db"`
}

type Migration struct {
	Version string `json:"version"`
	Dirty   bool   `json:"dirty"`
}

func (m Migration) TableName() string {
	return "schema_migrations"
}
