package server

import (
	"context"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(handler http.Handler, address string) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    address,
			Handler: handler,
		},
	}
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
