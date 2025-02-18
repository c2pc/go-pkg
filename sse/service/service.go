package service

import (
	"context"
	"sync"

	"github.com/c2pc/go-pkg/v2/sse/model"
)

type Client struct {
	ID      int
	Channel chan model.Data
}

type SSEManager struct {
	mu          sync.RWMutex
	clients     map[int]chan model.Data
	Broadcast   chan model.Data
	newClient   chan Client
	closeClient chan Client
	done        chan struct{}
	LenChan     int
}

func NewSSEManager(lenChan int) *SSEManager {
	mgr := &SSEManager{
		clients:     make(map[int]chan model.Data),
		Broadcast:   make(chan model.Data, lenChan),
		newClient:   make(chan Client, lenChan),
		closeClient: make(chan Client, lenChan),
		done:        make(chan struct{}),
		LenChan:     lenChan,
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

		case msg := <-mgr.Broadcast:
			mgr.mu.RLock()
			if msg.To != nil {
				if ch, ok := mgr.clients[*msg.To]; ok {
					ch <- msg
				}
			} else {
				for _, ch := range mgr.clients {
					ch <- msg
				}
			}
			mgr.mu.RUnlock()
		}
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

func (mgr *SSEManager) SendMessage(ctx context.Context, m model.Message) error {
	data := model.Data{
		Message:       m.Message,
		MessageType:   m.Type,
		MessageAction: m.Action,
		From:          m.From,
		To:            m.To,
	}
	if m.Topic != nil {
		data.Topic = string(*m.Topic)
	}

	if m.PushType != nil {
		data.PushType = *m.PushType
	}

	select {
	case mgr.Broadcast <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
