package main

import (
	"sync"
)

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

type Client struct {
	ID      int
	Channel chan Data
}

type SSEManager struct {
	mu          sync.RWMutex
	clients     map[int]chan Data
	broadcast   chan Data
	newClient   chan Client
	closeClient chan Client
	done        chan struct{}
	lenChan     int
}

func NewSSEManager(lenChan int) *SSEManager {
	mgr := &SSEManager{
		clients:     make(map[int]chan Data),
		broadcast:   make(chan Data, lenChan),
		newClient:   make(chan Client, lenChan),
		closeClient: make(chan Client, lenChan),
		done:        make(chan struct{}),
		lenChan:     lenChan,
	}

	go mgr.run()

	return mgr
}

func (mgr *SSEManager) run() {
	for {
		select {
		case <-mgr.done:
			return

		case client := <-mgr.newClient:
			mgr.mu.Lock()
			mgr.clients[client.ID] = client.Channel
			mgr.mu.Unlock()

		case client := <-mgr.closeClient:
			mgr.mu.Lock()
			if _, ok := mgr.clients[client.ID]; ok {
				delete(mgr.clients, client.ID)
				close(client.Channel)
			}
			mgr.mu.Unlock()

		case msg := <-mgr.broadcast:
			mgr.mu.RLock()
			for _, ch := range mgr.clients {
				ch <- msg
			}
			mgr.mu.RUnlock()
		}
	}
}

func (mgr *SSEManager) Broadcast(d Data) {
	select {
	case mgr.broadcast <- d:
	default:
		mgr.broadcast <- d
	}
}

func (mgr *SSEManager) RegisterClient(c Client) {
	mgr.newClient <- c
}

func (mgr *SSEManager) UnregisterClient(c Client) {
	mgr.closeClient <- c
}

func (mgr *SSEManager) Shutdown() {
	close(mgr.done)
}
