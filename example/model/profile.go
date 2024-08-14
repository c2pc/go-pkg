package model

type Profile struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`

	UserID int `json:"user_id"`
}

func (m Profile) TableName() string {
	return "auth_profiles"
}
