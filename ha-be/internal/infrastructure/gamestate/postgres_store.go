package gamestate

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

// gameStateRecord is the GORM model for the game_states table.
type gameStateRecord struct {
	Slug    string `gorm:"primaryKey"`
	Payload []byte `gorm:"type:jsonb"`
}

func (gameStateRecord) TableName() string { return "game_states" }

// PostgresGameStateStore persists GameUI as a JSONB column in Postgres.
type PostgresGameStateStore struct {
	db *gorm.DB
}

// NewPostgresGameStateStore creates a store backed by the provided gorm.DB.
func NewPostgresGameStateStore(db *gorm.DB) *PostgresGameStateStore {
	return &PostgresGameStateStore{db: db}
}

// AutoMigrate creates or updates the game_states table.
func (s *PostgresGameStateStore) AutoMigrate() error {
	return s.db.AutoMigrate(&gameStateRecord{})
}

func (s *PostgresGameStateStore) Save(ctx context.Context, slug string, state *game.GameUI) error {
	if s.db == nil {
		return fmt.Errorf("postgres game state store: db is nil")
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("postgres game state store marshal: %w", err)
	}
	rec := gameStateRecord{Slug: slug, Payload: payload}
	return s.db.WithContext(ctx).
		Save(&rec).Error
}

func (s *PostgresGameStateStore) Load(ctx context.Context, slug string) (*game.GameUI, error) {
	if s.db == nil {
		return nil, fmt.Errorf("postgres game state store: db is nil")
	}
	var rec gameStateRecord
	result := s.db.WithContext(ctx).First(&rec, "slug = ?", slug)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("postgres game state store load: %w", result.Error)
	}
	var state game.GameUI
	if err := json.Unmarshal(rec.Payload, &state); err != nil {
		return nil, fmt.Errorf("postgres game state store unmarshal: %w", err)
	}
	return &state, nil
}

func (s *PostgresGameStateStore) Delete(ctx context.Context, slug string) error {
	if s.db == nil {
		return fmt.Errorf("postgres game state store: db is nil")
	}
	return s.db.WithContext(ctx).
		Delete(&gameStateRecord{}, "slug = ?", slug).Error
}
