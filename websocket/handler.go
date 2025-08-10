package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	http2 "net/http"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

var (
	ErrMaxCountSessions = apperr.New("max_count_sessions",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Превышено максимальное количество сессий", translator.EN: "Maximum number of sessions exceeded"}),
		apperr.WithCode(code.Unauthenticated))
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var (
	MaxMessageSize int64 = 512
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
	manager  *manager
	maxCount int
}

func newWebSocket(manager *manager, maxCount int) *handler {
	return &handler{manager: manager, maxCount: maxCount}
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

	sessionID := uuid.New().String()

	err := func() error {
		s.manager.mu.RLock()
		defer s.manager.mu.RUnlock()
		clientSessions, ok := s.manager.clients[userID]
		if ok {
			if len(clientSessions) >= s.maxCount {
				return ErrMaxCountSessions.WithErrorText(fmt.Sprintf("maximum number of clients reached: %d - %d", len(clientSessions), s.maxCount))
			}
		}
		return nil
	}()
	if err != nil {
		http.Response(c, err)
		return
	}

	cl := Client{
		ID:        userID,
		ch:        make(chan broadcast, s.manager.lenChan),
		sessionID: sessionID,
	}

	upgrader.CheckOrigin = func(r *http2.Request) bool { return true }

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	s.manager.registerClient(&cl)

	go s.writePump(c.Request.Context(), conn, &cl)
	go s.readPump(c.Request.Context(), conn, &cl)
}

func (s *handler) readPump(ctx context.Context, conn *ws.Conn, client *Client) {
	defer func() {
		s.manager.unregisterClient(client)
		_ = conn.Close()
	}()
	conn.SetReadLimit(MaxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { _ = conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		select {
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
					if logger.IsDebugEnabled(level.TEST) {
						logger.WarningfLog(ctx, "WS", "error: %v", err)
					}
				}
				break
			}
			message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
			s.manager.notifyNewMessage(client, message)
		}
	}
}

func (s *handler) writePump(ctx context.Context, conn *ws.Conn, client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()
	for {
		select {
		case message, ok := <-client.ch:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}

			m, tp, err := getContent(message)
			if err != nil {
				continue
			}

			w, err := conn.NextWriter(tp)
			if err != nil {
				return
			}

			_, _ = w.Write(m)

			n := len(client.ch)
			for i := 0; i < n; i++ {
				_, _ = w.Write(newline)
				m, tp, err = getContent(<-client.ch)
				if err != nil {
					continue
				}

				w, err = conn.NextWriter(tp)
				if err != nil {
					return
				}

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

func getContent(message broadcast) ([]byte, int, error) {
	var m []byte
	var err error
	var tp = ws.TextMessage

	switch message.ContentType {
	case 0:
		m, err = json.Marshal(message)
		if err != nil {
			return nil, 0, err
		}
	case ws.TextMessage:
		b, ok := message.Message.(string)
		if !ok {
			return nil, 0, errors.New("message is not a TextMessage")
		}
		m = []byte(b)
	case ws.BinaryMessage:
		b, ok := message.Message.([]byte)
		if !ok {
			return nil, 0, errors.New("message is not a TextMessage")
		}
		m = b
	}

	return m, tp, nil
}
