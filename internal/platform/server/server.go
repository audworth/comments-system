package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/audworth/comments-system/internal/application/comment"
	"github.com/audworth/comments-system/internal/application/post"
	"github.com/audworth/comments-system/internal/application/user"
	"github.com/audworth/comments-system/internal/config"
	"github.com/audworth/comments-system/internal/platform/logger"
	"github.com/audworth/comments-system/internal/transport/graph"
	"github.com/audworth/comments-system/internal/transport/graph/resolver"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

type Server struct {
	cfg config.Config
}

func New(cfg config.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Run(ctx context.Context) error {
	lg, err := logger.New(s.cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	// TODO: real deps
	postsSvc := post.NewService(nil)
	commsSvc := comment.NewService(nil, nil)
	usersSvc := user.NewService(nil)

	root := resolver.New(postsSvc, usersSvc, commsSvc)
	handler := graph.NewHandler(
		root,
		usersSvc,
		commsSvc,
		lg,
		graph.HandlerConfig{
			Local:           s.cfg.Env == config.LocalEnv,
			ComplexityLimit: s.cfg.QueryComplexityLimit,
		},
	)
	server := &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	lg.InfoContext(ctx, "HTTP server started", slog.String("address", server.Addr))
	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)

	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		lg.InfoContext(ctx, "shutting down HTTP server", slog.Duration("shutdown-timeout", shutdownTimeout))
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}

		lg.InfoContext(ctx, "graceful shutdown completed")
		return nil
	}
}
