package request

type AuthLoginRequest struct {
	Login    string `json:"login" binding:"required,max=255,dot_underscore_hyphen"`
	Password string `json:"password" binding:"required,max=255,spec_chars"`
	DeviceID int    `json:"device_id" binding:"required,device_id"`
}

type AuthRefreshRequest struct {
	Token    string `json:"token" binding:"required"`
	DeviceID int    `json:"device_id" binding:"required,device_id"`
}

type AuthLogoutRequest struct {
	Token string `json:"accessToken" binding:"required"`
}

type AuthUpdateAccountDataRequest struct {
	Login      *string `json:"login" binding:"omitempty,max=255,min=2,dot_underscore_hyphen"`
	FirstName  *string `json:"first_name" binding:"omitempty,max=255,min=2,dot_underscore_hyphen"`
	SecondName *string `json:"second_name" binding:"omitempty,len=0|min=2,max=255,dot_underscore_hyphen"`
	LastName   *string `json:"last_name" binding:"omitempty,len=0|min=2,max=255,dot_underscore_hyphen"`
	Password   *string `json:"password" binding:"omitempty,max=255,min=8"`
	Email      *string `json:"email" binding:"omitempty,len=0|email,max=255"`
	Phone      *string `json:"phone" binding:"omitempty,len=0|min=1,max=255"`
}
