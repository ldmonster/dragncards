package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context, url string) (*redis.Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("redis url empty")
	}

	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return client, nil
}
