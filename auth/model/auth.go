package model

type Token struct {
	Token        string  `json:"token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresAt    float64 `json:"expires_at"`
	TokenType    string  `json:"token_type"`
	UserID       int     `json:"user_id"`
}

type AuthToken struct {
	Auth Token `json:"auth"`
	User User  `json:"user"`
}
