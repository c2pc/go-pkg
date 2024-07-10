package rest

import (
	"context"
	"net/http"
	"time"
)

const (
	defaultHTTPRWTimeout          = 10 * time.Second
	defaultHTTPMaxHeaderMegabytes = 1
)

type Server struct {
	httpServer *http.Server
}

type Input struct {
	Host string
	Port string
}

func NewServer(input Input, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:           input.Host + ":" + input.Port,
			Handler:        handler,
			ReadTimeout:    defaultHTTPRWTimeout,
			WriteTimeout:   defaultHTTPRWTimeout,
			MaxHeaderBytes: defaultHTTPMaxHeaderMegabytes << 10,
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
