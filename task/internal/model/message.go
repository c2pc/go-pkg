package model

const (
	TaskMessageType = "tasks"
)

const (
	TaskStatusChangedMessageAction = "status-changed"
)

type TaskMessage struct {
	Status string `json:"status"`
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}
