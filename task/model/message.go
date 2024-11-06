package model

import (
	"sync"
)

type Error struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Message struct {
	mu      sync.Mutex `json:"-"`
	Count   int        `json:"count"`
	Success []string   `json:"success,omitempty"`
	Errors  []Error    `json:"errors,omitempty"`
	Error   *string    `json:"error,omitempty"`
	data    []byte     `json:"-"`
}

func NewMessage() *Message {
	return &Message{
		mu:    sync.Mutex{},
		Count: 0,
	}
}

func (m *Message) AddError(key string, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Errors == nil {
		m.Errors = []Error{}
	}
	m.Errors = append(m.Errors, Error{Key: key, Value: value})
}

func (m *Message) GetErrors() []Error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.Errors
}

func (m *Message) AddSuccess(idx string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Success == nil {
		m.Success = []string{}
	}
	m.Success = append(m.Success, idx)
}

func (m *Message) GetSuccesses() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Success
}

func (m *Message) SetCount(value int) {
	m.Count = value
}

func (m *Message) GetCount() int {
	return m.Count
}

func (m *Message) SetData(value []byte) {
	m.data = value
}

func (m *Message) GetData() []byte {
	return m.data
}

func (m *Message) SetError(err string) {
	m.Error = &err
}

func (m *Message) GetError() *string {
	return m.Error
}
