package websocket

import (
	"context"
	"sync"
)

const (
	NewClientListenerType   = "new-Client"
	CloseClientListenerType = "close-Client"
	NewMessageListenerType  = "new-message"
)

type Listener struct {
	Event    string
	Message  []byte
	ClientID int
}

type Client struct {
	ID   int
	send chan broadcast
}

type manager struct {
	mu         sync.RWMutex
	clients    map[int]chan broadcast
	broadcast  chan broadcast
	register   chan *Client
	unregister chan *Client
	done       chan struct{}
	lenChan    int
	listeners  []chan Listener
}

func newWebSocketManager(lenChan int) *manager {
	mgr := &manager{
		clients:    make(map[int]chan broadcast),
		broadcast:  make(chan broadcast, lenChan),
		register:   make(chan *Client, lenChan),
		unregister: make(chan *Client, lenChan),
		done:       make(chan struct{}),
		lenChan:    lenChan,
		listeners:  []chan Listener{},
	}

	go mgr.run()

	return mgr
}

func (mgr *manager) run() {
	for {
		select {
		case <-mgr.done:
			return

		case client := <-mgr.register:
			mgr.mu.Lock()
			mgr.clients[client.ID] = client.send
			go mgr.notifyNewClient(client)
			mgr.mu.Unlock()

		case client := <-mgr.unregister:
			mgr.mu.Lock()
			_, ok := mgr.clients[client.ID]
			if ok {
				delete(mgr.clients, client.ID)
				close(client.send)
				go mgr.notifyCloseClient(client)
			}
			mgr.mu.Unlock()

		case msg := <-mgr.broadcast:
			mgr.mu.RLock()
			if msg.To != nil && len(msg.To) > 0 {
				for _, to := range msg.To {
					if ch, ok := mgr.clients[to]; ok {
						ch <- msg
					}
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

func (mgr *manager) registerClient(c *Client) {
	mgr.register <- c
}

func (mgr *manager) unregisterClient(c *Client) {
	mgr.unregister <- c
}

func (mgr *manager) shutdown() {
	close(mgr.done)
}

func (mgr *manager) sendMessage(ctx context.Context, m Message) error {
	d := broadcast{
		Message:       m.Message,
		MessageType:   m.Type,
		MessageAction: m.Action,
		From:          m.From,
		To:            m.To,
	}

	select {
	case mgr.broadcast <- d:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (mgr *manager) notifyNewClient(client *Client) {
	for _, ch := range mgr.listeners {
		ch <- Listener{Event: NewClientListenerType, ClientID: client.ID}
	}
}

func (mgr *manager) notifyCloseClient(client *Client) {
	for _, ch := range mgr.listeners {
		ch <- Listener{Event: CloseClientListenerType, ClientID: client.ID}
	}
}

func (mgr *manager) notifyNewMessage(client *Client, msg []byte) {
	for _, ch := range mgr.listeners {
		ch <- Listener{Event: NewMessageListenerType, ClientID: client.ID, Message: msg}
	}
}

func (mgr *manager) registerListener() <-chan Listener {
	ch := make(chan Listener)
	mgr.listeners = append(mgr.listeners, ch)
	return ch
}
