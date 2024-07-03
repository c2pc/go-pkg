package model

type Permission struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (m Permission) TableName() string {
	return "permissions"
}
