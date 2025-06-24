package websocket

type Message struct {
	Type        string      `json:"type"`
	Action      string      `json:"action"`
	Message     interface{} `json:"message"`
	From        *int        `json:"from,omitempty"`
	To          []int       `json:"to,omitempty"`
	ContentType int         `json:"-"`
}

type broadcast struct {
	Message       interface{} `json:"message"`
	MessageType   string      `json:"message_type"`
	MessageAction string      `json:"message_action"`
	From          *int        `json:"-"`
	To            []int       `json:"-"`
	ContentType   int         `json:"-"`
}
