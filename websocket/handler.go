package websocket

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type handler struct {
	manager *manager
}

func newWebSocket(manager *manager) *handler {
	return &handler{manager: manager}
}

func (s *handler) Init(api *gin.RouterGroup) {
	api.GET("stream", s.Stream)
}

func (s *handler) Stream(c *gin.Context) {
	userID, ok := mcontext.GetOpUserID(c.Request.Context())
	if !ok {
		http.Response(c, apperr.ErrUnauthenticated.WithErrorText("operation user id is empty"))
		return
	}

	cl := Client{
		ID:   userID,
		send: make(chan broadcast, s.manager.lenChan),
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	s.manager.registerClient(&cl)

	go s.writePump(conn, &cl)
	go s.readPump(conn, &cl)
}

func (s *handler) readPump(conn *ws.Conn, client *Client) {
	defer func() {
		s.manager.unregisterClient(client)
		_ = conn.Close()
	}()
	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { _ = conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		s.manager.notifyNewMessage(client, message)
	}
}

func (s *handler) writePump(conn *ws.Conn, client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.send:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(ws.TextMessage)
			if err != nil {
				return
			}

			m, _ := json.Marshal(message)
			_, _ = w.Write(m)

			n := len(client.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write(newline)
				m, _ = json.Marshal(<-client.send)
				_, _ = w.Write(m)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
