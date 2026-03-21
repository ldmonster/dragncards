package game

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/ldmonster/dragncards/ha-be/internal/application/plugin"
	domaingame "github.com/ldmonster/dragncards/ha-be/internal/domain/game"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/game/evaluate"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/gamestate"
)

// GameService manages room goroutines and routes game actions.
type GameService struct {
	roomService   *room.RoomService
	registry      *RoomRegistry
	stateStore    gamestate.GameStateStore // optional; nil = in-memory only
	pluginService *plugin.Service          // optional plugin DSL info source
	log           *slog.Logger
}

// NewGameService creates a GameService with optional GameStateStore and optional plugin service.
// Pass nil for stateStore to keep game state purely in-memory.
func NewGameService(roomService *room.RoomService, registry *RoomRegistry, stateStore gamestate.GameStateStore) *GameService {
	return NewGameServiceWithPlugin(roomService, registry, stateStore, nil)
}

func NewGameServiceWithPlugin(roomService *room.RoomService, registry *RoomRegistry, stateStore gamestate.GameStateStore, pluginService *plugin.Service) *GameService {
	return &GameService{
		roomService:   roomService,
		registry:      registry,
		stateStore:    stateStore,
		pluginService: pluginService,
		log:           slog.Default(),
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

			gameRoom.mu.Lock()
			gameRoom.State.AddAction(payload)

			if err := s.applyGameAction(gameRoom, payload); err != nil {
				s.log.Error("game_service: failed to apply game action",
					"slug", gameRoom.Slug,
					"error", err)
			}

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
			gameRoom.mu.Unlock()
		}
	}
}

func (s *GameService) applyGameAction(gameRoom *GameRoom, payload []byte) error {
	var env struct {
		Actor   string          `json:"actor,omitempty"`
		Action  string          `json:"action,omitempty"`
		Options json.RawMessage `json:"options,omitempty"`
	}
	if err := json.Unmarshal(payload, &env); err != nil {
		return err
	}

	switch strings.ToLower(env.Action) {
	case "set_game":
		var opts struct {
			Game *domaingame.GameUI `json:"game"`
		}
		if err := json.Unmarshal(env.Options, &opts); err != nil {
			return err
		}
		if opts.Game == nil {
			return errors.New("set_game option requires game")
		}
		gameRoom.State = opts.Game
		return nil

	case "evaluate":
		var opts struct {
			ActionList []json.RawMessage `json:"action_list"`
		}
		if err := json.Unmarshal(env.Options, &opts); err != nil {
			return err
		}
		for _, item := range opts.ActionList {
			if err := s.executeEvaluateItem(gameRoom, item); err != nil {
				return err
			}
		}
		return nil

	default:
		// no action semantics for other types yet
		return nil
	}
}

func (s *GameService) executeEvaluateItem(gameRoom *GameRoom, raw json.RawMessage) error {
	var val any
	if err := json.Unmarshal(raw, &val); err != nil {
		return nil
	}

	// Handle embedded game_action event in action_list items.
	if m, ok := val.(map[string]any); ok {
		if t, ok := m["type"].(string); ok && strings.EqualFold(t, "game_action") {
			nested := struct {
				Actor   string          `json:"actor,omitempty"`
				Action  string          `json:"action,omitempty"`
				Options json.RawMessage `json:"options,omitempty"`
			}{
				Actor: "",
			}
			if action, ok := m["action"].(string); ok {
				nested.Action = action
			}
			if options, ok := m["options"]; ok {
				if b, err := json.Marshal(options); err == nil {
					nested.Options = b
				}
			}
			nestedBytes, _ := json.Marshal(nested)
			return s.applyGameAction(gameRoom, nestedBytes)
		}
	}

	ctx := evaluate.NewEvalContext(gameRoom.State)
	if s.pluginService != nil {
		if pluginCards, err := s.pluginService.LoadCardsByPlugin(); err == nil {
			ctx.Vars["plugin_cards"] = pluginCards
		}
	}
	_, err := evaluate.EvaluateExpression(ctx, nil, val)
	return err
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

func (s *GameService) SetSeat(slug, playerID, seat string) error {
	room := s.registry.Get(slug)
	if room == nil {
		return errors.New("room not found")
	}
	room.mu.Lock()
	room.State.SetSeat(playerID, seat)
	room.mu.Unlock()

	if s.stateStore != nil {
		room.mu.RLock()
		state := room.State
		room.mu.RUnlock()
		if err := s.stateStore.Save(context.Background(), slug, state); err != nil {
			s.log.Error("game_service: failed to save state after set_seat",
				"slug", slug, "error", err)
		}
	}
	return nil
}

func (s *GameService) SetSpectator(slug, playerID string, spectator bool) error {
	room := s.registry.Get(slug)
	if room == nil {
		return errors.New("room not found")
	}
	room.mu.Lock()
	room.State.SetSpectator(playerID, spectator)
	room.mu.Unlock()

	if s.stateStore != nil {
		room.mu.RLock()
		state := room.State
		room.mu.RUnlock()
		if err := s.stateStore.Save(context.Background(), slug, state); err != nil {
			s.log.Error("game_service: failed to save state after set_spectator",
				"slug", slug, "error", err)
		}
	}
	return nil
}

// GetGameUI returns the current in-memory state for a room.
func (s *GameService) GetGameUI(slug string) (*domaingame.GameUI, error) {
	room := s.registry.Get(slug)
	if room == nil {
		return nil, errors.New("room not found")
	}
	return room.getState(), nil
}

// ResetGame clears the action log for a room.
func (s *GameService) ResetGame(ctx context.Context, slug string) error {
	room := s.registry.Get(slug)
	if room == nil {
		return errors.New("room not found")
	}
	room.mu.Lock()
	room.State.ResetActions()
	room.State.ResetSeats()
	room.State.ResetSpectators()
	room.mu.Unlock()

	if s.stateStore != nil {
		room.mu.RLock()
		state := room.State
		room.mu.RUnlock()
		if err := s.stateStore.Save(ctx, slug, state); err != nil {
			s.log.Error("game_service: failed to save state after reset",
				"slug", slug, "error", err)
		}
	}
	return nil
}
