package model

import (
	"fmt"
	"time"
)

const (
	StatusPending = "pending"
	StatusRunning = "running"
	StatusStopped = "stopped"
	StatusFailed  = "failed"
	StatusSuccess = "success"
	StatusWarning = "warning"
)

var Statuses = map[string]string{
	StatusPending: StatusPending,
	StatusRunning: StatusRunning,
	StatusStopped: StatusStopped,
	StatusFailed:  StatusFailed,
	StatusSuccess: StatusSuccess,
	StatusWarning: StatusWarning,
}

type User struct {
	ID         int     `json:"id"`
	Login      string  `json:"login"`
	FirstName  string  `json:"first_name"`
	SecondName *string `json:"second_name"`
	LastName   *string `json:"last_name"`
}

func (m User) TableName() string {
	return "auth_users"
}

type Task struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	UserID    int       `json:"user_id"`
	Status    string    `json:"status"`
	Output    []byte    `json:"output"`
	Input     []byte    `json:"input"`
	FileSize  *int64    `json:"file_size" gorm:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *User `json:"user"`
}

func (m Task) TableName() string {
	return "auth_tasks"
}

func (m Task) FilePath() string {
	return fmt.Sprintf("media/tasks/%s_task_%d.csv", m.Type, m.ID)
}
