package websocket

type Message struct {
	Type          string      `json:"type"`
	Action        string      `json:"action"`
	Message       interface{} `json:"message"`
	From          *int64      `json:"from,omitempty"`
	To            []int64     `json:"to,omitempty"`
	ToSession     *string     `json:"to_session,omitempty"`
	ExceptSession *string     `json:"except_session,omitempty"`
	ContentType   int         `json:"-"`
}

type broadcast struct {
	Message       interface{} `json:"message"`
	MessageType   string      `json:"message_type"`
	MessageAction string      `json:"message_action"`
	From          *int64      `json:"-"`
	To            []int64     `json:"-"`
	ToSession     *string     `json:"-"`
	ExceptSession *string     `json:"-"`
	ContentType   int         `json:"-"`
}
