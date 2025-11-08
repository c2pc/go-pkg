package profile

type Profile struct {
	ID      int    `json:"id"`
	Age     *int   `json:"age"`
	Height  *int   `json:"height"`
	Address string `json:"address"`

	UserID int64 `json:"user_id"`
}

func (m Profile) TableName() string {
	return "auth_profiles"
}

func (m Profile) GetUserId() int64 {
	return m.UserID
}
