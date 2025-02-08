package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/tatucosmin/hotel-system/config"
	"github.com/tatucosmin/hotel-system/store"
)

type Server struct {
	Config     *config.Config
	logger     *slog.Logger
	store      *store.Store
	jwtManager *JwtManager
}

func New(cfg *config.Config, logger *slog.Logger, store *store.Store, jwtManager *JwtManager) *Server {
	return &Server{
		Config:     cfg,
		logger:     logger,
		store:      store,
		jwtManager: jwtManager,
	}
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", s.ping)
	mux.HandleFunc("POST /api/auth/signup", s.signUpHandler())
	mux.HandleFunc("POST /api/auth/signin", s.signInHandler())
	mux.HandleFunc("POST /api/auth/refresh", s.refreshTokenHandler())

	middlewareLogger := NewLoggerMiddleware(s.logger)
	middlewareAuth := NewAuthMiddleware(s.jwtManager, s.store.User)

	middleware := middlewareLogger(middlewareAuth(mux))

	server := &http.Server{
		Addr:    net.JoinHostPort(s.Config.ServerHost, s.Config.ServerPort),
		Handler: middleware,
	}

	go func() {
		s.logger.Info("server is running on", "port", s.Config.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server failed to serve", "error", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()

		closeCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := server.Shutdown(closeCtx); err != nil {
			s.logger.Error("server failed to shutdown", "error", err)
		}
	}()

	wg.Wait()

	return nil

}
