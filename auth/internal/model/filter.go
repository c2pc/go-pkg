package model

type Filter struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	DeviceID int    `json:"device_id"`
	Endpoint string `json:"endpoint"`
	Name     string `json:"name"`
	Value    []byte `json:"value"`

	User *User `json:"user"`
}

func (m Filter) TableName() string {
	return "auth_filters"
}
