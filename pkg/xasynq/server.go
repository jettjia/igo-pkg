package xasynq

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
)

// Server is a wrapper around asynq.Server providing integrated features.
type Server struct {
	srv *asynq.Server
	mux *asynq.ServeMux
	cfg ServerConfig
}

// NewServer creates a new consumer server.
func NewServer(redisCfg RedisConfig, serverCfg ServerConfig) *Server {
	if serverCfg.Config == nil {
		serverCfg = DefaultServerConfig()
	}
	srv := asynq.NewServer(redisCfg.GetAsynqRedisConnOpt(), *serverCfg.Config)
	return &Server{
		srv: srv,
		mux: asynq.NewServeMux(),
		cfg: serverCfg,
	}
}

// Mux returns the underlying ServeMux to register handlers.
func (s *Server) Mux() *asynq.ServeMux {
	return s.mux
}

// Use adds middleware to the server's handler chain.
func (s *Server) Use(middlewares ...asynq.MiddlewareFunc) {
	s.mux.Use(middlewares...)
}

// Register a task processor
func (s *Server) Register(typeName string, handler asynq.Handler) {
	s.mux.Handle(typeName, handler)
}

// RegisterFunc a task handler function
func (s *Server) RegisterFunc(typeName string, handlerFunc asynq.HandlerFunc) {
	s.mux.HandleFunc(typeName, handlerFunc)
}

// Run runs the asynq server in a separate goroutine
func (s *Server) Run() {
	go func() {
		if err := s.srv.Run(s.mux); err != nil {
			panic(fmt.Sprintf("could not run asynq server: %v", err))
		}
	}()
}

// Shutdown the server.
func (s *Server) Shutdown() {
	if s == nil || s.srv == nil {
		return
	}
	s.srv.Shutdown()
}

// WaitShutdown for interrupt signals for graceful shutdown the server.
func (s *Server) WaitShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.Shutdown()
}
