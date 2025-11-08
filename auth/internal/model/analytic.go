package model

import (
	"time"
)

type Analytic struct {
	ID           int64     `json:"id" gorm:"primary_key"`
	OperationID  string    `json:"operation_id"`
	Path         string    `json:"path"`
	Method       string    `json:"method"`
	StatusCode   int       `json:"status_code"`
	ClientIP     string    `json:"client_ip"`
	RequestBody  []byte    `json:"request_body" gorm:"type:bytea;null"`
	ResponseBody []byte    `json:"response_body" gorm:"type:bytea;null"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       *int64    `json:"user_id"`
	Login        *string   `json:"login"`
	Name         *string   `json:"name"`
	Error        *string   `json:"error"`
	Action       *string   `json:"action"`

	User *User `json:"user"`
}

func (a Analytic) TableName() string {
	return "auth_analytics_admins"
}
