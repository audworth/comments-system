package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context, url string) (*goredis.Client, error) {
	opts, err := goredis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	cl := goredis.NewClient(opts)
	if err := cl.Ping(ctx).Err(); err != nil {
		_ = cl.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return cl, nil
}
