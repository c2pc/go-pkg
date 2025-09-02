package websocket

type Message struct {
	Type        string      `json:"type"`
	Action      string      `json:"action"`
	Message     interface{} `json:"message"`
	From        *int        `json:"from,omitempty"`
	To          []int       `json:"to,omitempty"`
	ToSessionID *string     `json:"to_session_id,omitempty"`
	ContentType int         `json:"-"`
}

type broadcast struct {
	Message       interface{} `json:"message"`
	MessageType   string      `json:"message_type"`
	MessageAction string      `json:"message_action"`
	From          *int        `json:"-"`
	To            []int       `json:"-"`
	ToSessionID   *string     `json:"-"`
	ContentType   int         `json:"-"`
}
