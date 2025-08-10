package websocket

import (
	"context"
	"sync"
)

const (
	NewClientListenerType   = "new-client"
	CloseClientListenerType = "close-client"
	NewMessageListenerType  = "new-message"
)

type Listener struct {
	Event    string
	Message  []byte
	ClientID int
}

type Client struct {
	ID        int
	ch        chan broadcast
	sessionID string
}

type manager struct {
	mu         sync.RWMutex
	clients    map[int]map[string]*Client
	broadcast  chan broadcast
	register   chan *Client
	unregister chan *Client
	done       chan struct{}
	lenChan    int
	listeners  []chan Listener
}

func newWebSocketManager(lenChan int) *manager {
	mgr := &manager{
		clients:    make(map[int]map[string]*Client),
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
			func(cl *Client) {
				mgr.mu.Lock()
				defer mgr.mu.Unlock()
				_, ok := mgr.clients[cl.ID]
				if !ok {
					mgr.clients[cl.ID] = make(map[string]*Client)
					mgr.clients[cl.ID][cl.sessionID] = cl
					go mgr.notifyNewClient(cl)
				} else {
					mgr.clients[cl.ID][cl.sessionID] = cl
				}
			}(client)
		case client := <-mgr.unregister:
			func(cl *Client) {
				mgr.mu.Lock()
				defer mgr.mu.Unlock()

				clientSessions, ok := mgr.clients[cl.ID]
				if ok {
					if session, ok := clientSessions[cl.sessionID]; ok {
						close(session.ch)
						delete(clientSessions, cl.sessionID)
					}
					if len(clientSessions) == 0 {
						delete(mgr.clients, cl.ID)
						go mgr.notifyCloseClient(cl)
					}
				}
			}(client)
		case msg := <-mgr.broadcast:
			func(msg broadcast) {
				mgr.mu.RLock()
				defer mgr.mu.RUnlock()

				if msg.To != nil && len(msg.To) > 0 {
					for _, to := range msg.To {
						if clientSessions, ok := mgr.clients[to]; ok {
							for _, session := range clientSessions {
								session.ch <- msg
							}
						}
					}
				} else {
					for _, clientSessions := range mgr.clients {
						for _, session := range clientSessions {
							session.ch <- msg
						}
					}
				}
			}(msg)
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
	for _, clientSessions := range mgr.clients {
		for _, session := range clientSessions {
			close(session.ch)
		}
	}
	close(mgr.done)
}

func (mgr *manager) sendMessage(ctx context.Context, m Message) error {
	d := broadcast{
		Message:       m.Message,
		MessageType:   m.Type,
		MessageAction: m.Action,
		From:          m.From,
		To:            m.To,
		ContentType:   m.ContentType,
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
