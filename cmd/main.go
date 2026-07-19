package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/audworth/comments-system/internal/config"
	"github.com/audworth/comments-system/internal/platform/server"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatalf("error when running: %v", err)
	}
}

func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}

	app := server.New(cfg)
	return app.Run(ctx)
}
