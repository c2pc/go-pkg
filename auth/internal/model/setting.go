package model

type Setting struct {
	UserID   int    `json:"user_id"`
	DeviceID int    `json:"device_id"`
	Settings []byte `json:"settings"`

	User *User `json:"user"`
}

func (m Setting) TableName() string {
	return "auth_settings"
}
