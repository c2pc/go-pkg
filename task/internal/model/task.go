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
)

var Statuses = map[string]string{
	StatusPending: StatusPending,
	StatusRunning: StatusRunning,
	StatusStopped: StatusStopped,
	StatusFailed:  StatusFailed,
	StatusSuccess: StatusSuccess,
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
}

func (m Task) TableName() string {
	return "auth_tasks"
}

func (m Task) FilePath() string {
	return fmt.Sprintf("media/tasks/%s_task_%d.csv", m.Type, m.ID)
}
