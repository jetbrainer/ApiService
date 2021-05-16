package http

import (
	"context"

	"github.com/JetBrainer/ApiService/src/config"
	"github.com/JetBrainer/ApiService/src/manager"
)

type Option func(server *HTTPServer)

func WithContext(ctx context.Context) Option {
	return func(s *HTTPServer) {
		s.ctx = ctx
	}
}

func WithConfig(server *config.Server) Option {
	return func(s *HTTPServer) {
		s.addr = server.HTTPAddr
		if server.TLS != nil {
			s.certFile = server.TLS.CertFile
			s.keyFile = server.TLS.KeyFile
		}
	}
}

func WithRequestManager(req manager.Requester) Option {
	return func(server *HTTPServer) {
		server.requestMan = req
	}
}
