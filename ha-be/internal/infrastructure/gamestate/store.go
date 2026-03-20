package gamestate

import (
	"context"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

// GameStateStore is the pluggable persistence backend for game rooms.
// When REDIS_URL is set in config, a Redis-backed store is used; otherwise
// the Postgres-backed store persists game state as JSON.
type GameStateStore interface {
	// Save persists the full GameUI snapshot for the given room slug.
	Save(ctx context.Context, slug string, state *game.GameUI) error
	// Load retrieves the latest GameUI snapshot for the given room slug.
	// Returns (nil, nil) when no snapshot exists yet.
	Load(ctx context.Context, slug string) (*game.GameUI, error)
	// Delete removes the persisted state for the given room slug.
	Delete(ctx context.Context, slug string) error
}
