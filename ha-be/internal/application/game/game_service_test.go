package game

import (
	"context"
	"testing"
	"time"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/persistence"
)

func TestGameServiceCreateAndSendAction(t *testing.T) {
	roomRepo := persistence.NewInMemoryRoomRepository()
	roomSvc := room.NewService(roomRepo)
	registry := NewRoomRegistry()
	gameSvc := NewGameService(roomSvc, registry, nil)

	ctx := context.Background()
	gameUI, err := gameSvc.CreateGame(ctx, "test-game", "owner-1")
	if err != nil {
		t.Fatalf("CreateGame failed: %v", err)
	}
	if gameUI.Slug == "" {
		t.Fatalf("expected non-empty slug")
	}

	payload := []byte(`{"event":"move"}`)
	if err := gameSvc.SendAction(gameUI.Slug, payload); err != nil {
		t.Fatalf("SendAction failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	roomActions, err := roomSvc.ListActions(gameUI.Slug)
	if err != nil {
		t.Fatalf("RoomService list actions error: %v", err)
	}
	if len(roomActions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(roomActions))
	}
}
