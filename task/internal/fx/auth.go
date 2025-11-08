package fx

import (
	"sync/atomic"
)

type ConsumerHolder[T any] struct{ v atomic.Value }

func NewConsumersHolder[T any](consumers map[string]T) *ConsumerHolder[T] {
	h := &ConsumerHolder[T]{}
	h.Reload(consumers)
	return h
}

func (h *ConsumerHolder[T]) Get() map[string]T {
	return h.v.Load().(map[string]T)
}

func (h *ConsumerHolder[T]) Reload(consumers map[string]T) {
	h.v.Store(consumers)
	return
}
