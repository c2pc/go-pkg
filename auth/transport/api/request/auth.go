package request

type AuthLoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
	DeviceID int    `json:"device_id" binding:"required,device_id"`
}

type AuthRefreshRequest struct {
	Token    string `json:"token" binding:"required"`
	DeviceID int    `json:"device_id" binding:"required,device_id"`
}

type AuthLogoutRequest struct {
	Token string `json:"token" binding:"required"`
}
