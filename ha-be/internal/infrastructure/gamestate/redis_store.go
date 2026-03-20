package gamestate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
	"github.com/redis/go-redis/v9"
)

const redisKeyPrefix = "gamestate:"

// RedisGameStateStore persists GameUI as a JSON string in Redis.
// The key format is "gamestate:<slug>".
type RedisGameStateStore struct {
	client *redis.Client
}

// NewRedisGameStateStore creates a store backed by the provided Redis client.
func NewRedisGameStateStore(client *redis.Client) *RedisGameStateStore {
	return &RedisGameStateStore{client: client}
}

func (s *RedisGameStateStore) Save(ctx context.Context, slug string, state *game.GameUI) error {
	if s.client == nil {
		return fmt.Errorf("redis game state store: client is nil")
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("redis game state store marshal: %w", err)
	}
	return s.client.Set(ctx, redisKeyPrefix+slug, payload, 0).Err()
}

func (s *RedisGameStateStore) Load(ctx context.Context, slug string) (*game.GameUI, error) {
	if s.client == nil {
		return nil, fmt.Errorf("redis game state store: client is nil")
	}
	payload, err := s.client.Get(ctx, redisKeyPrefix+slug).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("redis game state store get: %w", err)
	}
	var state game.GameUI
	if err := json.Unmarshal(payload, &state); err != nil {
		return nil, fmt.Errorf("redis game state store unmarshal: %w", err)
	}
	return &state, nil
}

func (s *RedisGameStateStore) Delete(ctx context.Context, slug string) error {
	if s.client == nil {
		return fmt.Errorf("redis game state store: client is nil")
	}
	return s.client.Del(ctx, redisKeyPrefix+slug).Err()
}
