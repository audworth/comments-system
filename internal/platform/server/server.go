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
	"github.com/audworth/comments-system/internal/platform/db/inmem"
	"github.com/audworth/comments-system/internal/platform/db/postgres"
	"github.com/audworth/comments-system/internal/platform/logger"
	"github.com/audworth/comments-system/internal/platform/redis"
	"github.com/audworth/comments-system/internal/storage/mem"
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

	redisClient, err := redis.NewClient(ctx, s.cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("create redis: %w", err)
	}
	defer func() {
		_ = redisClient.Close()
	}()

	closeServices, err := s.initServices(ctx, redisClient, lg)
	if err != nil {
		return err
	}
	defer closeServices()

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

	lg.InfoContext(ctx, "server started")
	lg.InfoContext(
		ctx,
		"config",
		slog.String("address", s.cfg.Addr),
		slog.String("env", s.cfg.Env),
		slog.String("log level", s.cfg.LogLevel),
		slog.String("storage type", string(s.cfg.Storage)),
	)
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

func (s *Server) initServices(ctx context.Context, redisClient *goredis.Client, logger *slog.Logger) (func(), error) {
	notif := notifier.NewNotifier(redisClient, logger)
	sub, err := notifier.NewSubscriber(ctx, redisClient, logger)
	if err != nil {
		return nil, fmt.Errorf("create comment subscriber: %w", err)
	}

	switch s.cfg.Storage {
	case config.StoragePostgres:
		pool, err := postgres.New(ctx, postgres.Config{URL: s.cfg.DB})
		if err != nil {
			sub.Close()
			return nil, fmt.Errorf("connect to postgres: %w", err)
		}

		postsRepo := pg.NewPostRepository(pool, logger)
		commentsRepo := pg.NewCommentsRepository(pool, logger)
		usersRepo := pg.NewUserRepository(pool, logger)

		s.posts = post.NewService(postsRepo, logger)
		s.comments = comment.NewService(commentsRepo, notif, sub, logger)
		s.users = user.NewService(usersRepo, logger)

		return func() {
			sub.Close()
			pool.Close()
		}, nil

	case config.StorageMemory:
		inMem := inmem.New()
		if s.cfg.SeedInMem {
			inMem.Seed()
		}

		postsRepo := mem.NewPostRepository(inMem, logger)
		commentsRepo := mem.NewCommentsRepository(inMem, logger)
		usersRepo := mem.NewUserRepository(inMem, logger)

		s.posts = post.NewService(postsRepo, logger)
		s.comments = comment.NewService(commentsRepo, notif, sub, logger)
		s.users = user.NewService(usersRepo, logger)

		return sub.Close, nil

	default:
		sub.Close()
		return nil, fmt.Errorf("unsupported storage type %q", s.cfg.Storage)
	}
}
