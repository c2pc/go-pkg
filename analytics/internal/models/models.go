package models

import (
	"time"
)

type Analytics struct {
	ID           uint      `json:"id" gorm:"primary_key"`
	OperationID  string    `json:"operation_id"`
	Path         string    `json:"path"`
	UserID       *int      `json:"user_id"`
	Method       string    `json:"method"`
	StatusCode   int       `json:"status_code"`
	ClientIP     string    `json:"client_ip"`
	RequestBody  []byte    `json:"request_body" gorm:"type:bytea;null"`
	ResponseBody []byte    `json:"response_body" gorm:"type:bytea;null"`
	CreatedAt    time.Time `json:"created_at"`
	FirstName    string    `json:"first_name"`
	SecondName   string    `json:"second_name"`
	LastName     string    `json:"last_name"`
	Duration     int64     `json:"duration"`

	User *User `json:"user"`
}

func (a Analytics) TableName() string {
	return "auth_analytics"
}

type User struct {
	ID         int    `json:"id"`
	Login      string `json:"login"`
	FirstName  string `json:"first_name"`
	SecondName string `json:"second_name"`
	LastName   string `json:"last_name"`
}

func (m User) TableName() string {
	return "auth_users"
}
