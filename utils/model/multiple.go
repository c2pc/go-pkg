package model

import "sync"

type IMultiple interface {
	IDs() []int
}

type Multiple struct {
	ids []int
	mu  *sync.Mutex
}

func NewMultiple() *Multiple {
	return &Multiple{
		ids: []int{},
		mu:  &sync.Mutex{},
	}
}

func (m *Multiple) AddID(delta int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ids = append(m.ids, delta)
}

func (m *Multiple) IDs() []int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ids
}
