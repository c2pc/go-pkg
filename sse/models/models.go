package models

type Topic string

type PushType string

const (
	PushTypeBackground PushType = "background"
	PushTypeAlert      PushType = "alert"
)

type Message struct {
	Type     string      `json:"type"`
	Action   string      `json:"action"`
	PushType *PushType   `json:"push_type,omitempty"`
	Topic    *Topic      `json:"topic,omitempty"`
	Message  interface{} `json:"message"`
	From     *int        `json:"from,omitempty"`
	To       *int        `json:"to,omitempty"`
}

type Data struct {
	PushType      PushType    `json:"push_type,omitempty"`
	Topic         string      `json:"topic,omitempty"`
	Message       interface{} `json:"message"`
	MessageType   string      `json:"message_type"`
	MessageAction string      `json:"message_action"`
	From          *int        `json:"-"`
	To            *int        `json:"-"`
}
