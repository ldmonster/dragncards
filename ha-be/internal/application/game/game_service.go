package game

import (
	"context"
	"errors"
	"log/slog"
	"time"

	domaingame "github.com/ldmonster/dragncards/ha-be/internal/domain/game"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/gamestate"
)

// GameService manages room goroutines and routes game actions.
type GameService struct {
	roomService *room.RoomService
	registry    *RoomRegistry
	stateStore  gamestate.GameStateStore // optional; nil = in-memory only
	log         *slog.Logger
}

// NewGameService creates a GameService with optional GameStateStore.
// Pass nil for stateStore to keep game state purely in-memory.
func NewGameService(roomService *room.RoomService, registry *RoomRegistry, stateStore gamestate.GameStateStore) *GameService {
	return &GameService{
		roomService: roomService,
		registry:    registry,
		stateStore:  stateStore,
		log:         slog.Default(),
	}
}

// CreateGame creates a new room and starts its goroutine.
func (s *GameService) CreateGame(ctx context.Context, slug, ownerID string) (*domaingame.GameUI, error) {
	if slug == "" || ownerID == "" {
		return nil, errors.New("slug and owner required")
	}

	// Create or fetch the room record.
	roomObj, err := s.roomService.Create(slug, ownerID)
	if err != nil {
		return nil, err
	}

	// Attempt to restore persisted state; fall back to fresh state.
	var gameUI *domaingame.GameUI
	if s.stateStore != nil {
		if loaded, loadErr := s.stateStore.Load(ctx, slug); loadErr == nil && loaded != nil {
			gameUI = loaded
		}
	}
	if gameUI == nil {
		gameUI = domaingame.NewGameUI(roomObj.Slug)
	}

	rs := &GameRoom{
		Slug:  roomObj.Slug,
		State: gameUI,
		In:    make(chan []byte, 64),
		done:  make(chan struct{}),
	}
	s.registry.Register(rs)
	go s.run(ctx, rs)

	return gameUI, nil
}

// run is the long-lived goroutine for a single game room.
// It exits when the room's done channel is closed OR when ctx is cancelled.
func (s *GameService) run(ctx context.Context, gameRoom *GameRoom) {
	defer s.registry.Unregister(gameRoom.Slug)

	for {
		select {
		case <-ctx.Done():
			return
		case <-gameRoom.Done():
			return
		case payload, ok := <-gameRoom.In:
			if !ok {
				return
			}
			gameRoom.State.AddAction(payload)

			// Persist action via room service.
			if err := s.roomService.AppendAction(gameRoom.Slug, payload); err != nil {
				s.log.Error("game_service: failed to append action",
					"slug", gameRoom.Slug,
					"error", err)
			}

			// Persist full state snapshot if a store is configured.
			if s.stateStore != nil {
				if err := s.stateStore.Save(ctx, gameRoom.Slug, gameRoom.State); err != nil {
					s.log.Error("game_service: failed to save state snapshot",
						"slug", gameRoom.Slug,
						"error", err)
				}
			}
		}
	}
}

// SendAction enqueues a raw action payload into the room's inbox.
func (s *GameService) SendAction(slug string, payload []byte) error {
	room := s.registry.Get(slug)
	if room == nil {
		return errors.New("room not found")
	}
	if len(payload) == 0 {
		return errors.New("payload required")
	}
	select {
	case room.In <- payload:
		return nil
	case <-time.After(1 * time.Second):
		return errors.New("timeout sending action")
	}
}

// GetGameUI returns the current in-memory state for a room.
func (s *GameService) GetGameUI(slug string) (*domaingame.GameUI, error) {
	room := s.registry.Get(slug)
	if room == nil {
		return nil, errors.New("room not found")
	}
	return room.State, nil
}

// ResetGame clears the action log for a room.
func (s *GameService) ResetGame(ctx context.Context, slug string) error {
	room := s.registry.Get(slug)
	if room == nil {
		return errors.New("room not found")
	}
	room.State.ResetActions()
	if s.stateStore != nil {
		if err := s.stateStore.Save(ctx, slug, room.State); err != nil {
			s.log.Error("game_service: failed to save state after reset",
				"slug", slug, "error", err)
		}
	}
	return nil
}
