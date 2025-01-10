package models

type Topic string

type PushType string

const (
	PushTypeBackground PushType = "background"
	PushTypeAlert      PushType = "alert"
)

type Message struct {
	PushType *PushType `json:"push_type,omitempty"`
	Topic    *Topic    `json:"topic,omitempty"`
	Message  string    `json:"message"`
	From     *int      `json:"from,omitempty"`
	To       *int      `json:"to,omitempty"`
}

type Data struct {
	PushType PushType `json:"push_type,omitempty"`
	Topic    string   `json:"topic,omitempty"`
	Message  string   `json:"message"`
	From     int      `json:"from"`
	To       int      `json:"to"`
}
