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
	"github.com/audworth/comments-system/internal/notifier"
	"github.com/audworth/comments-system/internal/platform/db"
	"github.com/audworth/comments-system/internal/platform/logger"
	"github.com/audworth/comments-system/internal/platform/redis"
	"github.com/audworth/comments-system/internal/storage/pg"
	"github.com/audworth/comments-system/internal/transport/graph"
	"github.com/audworth/comments-system/internal/transport/graph/resolver"
	goredis "github.com/redis/go-redis/v9"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

type Server struct {
	cfg      config.Config
	users    *user.Service
	posts    *post.Service
	comments *comment.Service
}

func New(cfg config.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Run(ctx context.Context) error {
	lg, err := logger.New(s.cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	redis, err := redis.NewClient(ctx, s.cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("create redis: %w", err)
	}

	close, err := s.initServices(ctx, redis, lg)
	if err != nil {
		return err
	}
	defer close()

	root := resolver.New(s.posts, s.users, s.comments)
	graphHandler := graph.NewHandler(
		root,
		s.users,
		s.comments,
		lg,
		graph.HandlerConfig{Local: s.cfg.Env == config.LocalEnv},
	)

	server := &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           graphHandler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	lg.InfoContext(ctx, "HTTP server started", slog.String("address", server.Addr))
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
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

		lg.InfoContext(ctx, "shutting down HTTP server", slog.Duration("shutdown_timeout", shutdownTimeout))
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

func (s *Server) initServices(ctx context.Context, redis *goredis.Client, logger *slog.Logger) (func(), error) {
	switch s.cfg.Storage {
	case config.StoragePostgres:
		pool, err := db.NewPostgres(ctx, db.Config{URL: s.cfg.DB})
		if err != nil {
			return nil, fmt.Errorf("connect to postgres: %w", err)
		}

		postsRepo := pg.NewPostRepository(pool, logger)
		commentsRepo := pg.NewCommentsRepository(pool, logger)
		usersRepo := pg.NewUserRepository(pool, logger)

		notif := notifier.NewNotifier(redis, logger)
		sub := notifier.NewSubscriber(redis, logger)

		s.posts = post.NewService(postsRepo)
		s.comments = comment.NewService(commentsRepo, notif, sub, logger)
		s.users = user.NewService(usersRepo)

		return func() {
			_ = redis.Close()
			pool.Close()
		}, nil

	case config.StorageMemory:
		return func() {
			_ = redis.Close()
		}, nil

	default:
		return nil, fmt.Errorf("unsupported storage type %q", s.cfg.Storage)
	}
}
