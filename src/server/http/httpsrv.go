package http

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JetBrainer/ApiService/src/manager"
	v1 "github.com/JetBrainer/ApiService/src/server/http/resources/v1"
)

const (
	readTimeout  = 10 * time.Second
	writeTimeout = 15 * time.Second

	shutdownTTL = 10 * time.Second
)

type HTTPServer struct {
	addr       string
	keyFile    string
	certFile   string
	ctx        context.Context
	requestMan manager.Requester
}

func NewHTTPServer(opts ...Option) *HTTPServer {
	var srv HTTPServer
	for _, opt := range opts {
		opt(&srv)
	}
	return &srv
}

// setupRouter there is routers that accepts requests from client
func (s *HTTPServer) setupRouter() http.Handler {
	r := http.NewServeMux()

	resource := v1.NewURLResource(s.requestMan)

	r.HandleFunc("/urls", resource.RequestLimiter(resource.URLHandler))

	return r
}

// Run our http server with graceful shutdown
func (s *HTTPServer) Run(cancel context.CancelFunc) {
	srv := s.NewServer()

	log.Printf("[INFO] Serving HTTP on %s", s.addr)

	srv.RegisterOnShutdown(cancel)
	srv.SetKeepAlivesEnabled(false)

	// Run server
	go func() {
		if s.keyFile != "" && s.certFile != "" {
			if err := srv.ListenAndServeTLS(s.certFile, s.keyFile); err != http.ErrServerClosed {
				log.Fatalf("HTTP server ListenAndServe: %v", err)
			}
		}
		if s.keyFile == "" && s.certFile == "" {
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatalf("HTTP server ListenAndServe: %v", err)
			}
		}
	}()

	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	<-signalChan
	log.Print("os.Interrupt - shutting down...\n")

	go func() {
		<-signalChan
		log.Fatal("os.Kill - terminating...\n")
	}()

	gracefullCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTTL)
	defer cancelShutdown()

	err := srv.Shutdown(gracefullCtx)
	if err != nil {
		log.Printf("shutdown error: %v\n", err)
		defer os.Exit(1)
		return
	}
	log.Printf("gracefully stopped\n")
	os.Exit(0)
}

func (s *HTTPServer) NewServer() *http.Server {
	return &http.Server{
		Addr:         s.addr,
		Handler:      s.setupRouter(),
		BaseContext:  func(_ net.Listener) context.Context { return s.ctx },
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
}
