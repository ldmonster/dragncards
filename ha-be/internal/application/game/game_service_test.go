package game

import (
	"context"
	"testing"
	"time"

	"github.com/ldmonster/dragncards/ha-be/internal/application/plugin"
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

func TestGameServiceSetGameAction(t *testing.T) {
	roomRepo := persistence.NewInMemoryRoomRepository()
	roomSvc := room.NewService(roomRepo)
	registry := NewRoomRegistry()
	gameSvc := NewGameService(roomSvc, registry, nil)

	ctx := context.Background()
	gameUI, err := gameSvc.CreateGame(ctx, "test-game-set", "owner-1")
	if err != nil {
		t.Fatalf("CreateGame failed: %v", err)
	}

	payload := []byte(`{"action":"set_game","options":{"game":{"slug":"test-game-set","players":["player1"]}}}`)
	if err := gameSvc.SendAction(gameUI.Slug, payload); err != nil {
		t.Fatalf("SendAction failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	updated, err := gameSvc.GetGameUI(gameUI.Slug)
	if err != nil {
		t.Fatalf("GetGameUI failed: %v", err)
	}
	if len(updated.Players) != 1 || updated.Players[0] != "player1" {
		t.Fatalf("expected player1 in game state, got %+v", updated.Players)
	}
}

func TestGameServiceEvaluateActionList(t *testing.T) {
	roomRepo := persistence.NewInMemoryRoomRepository()
	roomSvc := room.NewService(roomRepo)
	registry := NewRoomRegistry()
	pluginRepo := persistence.NewInMemoryPluginRepository()
	pluginSvc := plugin.NewService(pluginRepo)
	_, _ = pluginSvc.Create("test-plugin", true)
	_, _ = pluginSvc.CreateCustomCard("p-1", "card-1", "{\"foo\":\"bar\"}")
	gameSvc := NewGameServiceWithPlugin(roomSvc, registry, nil, pluginSvc)

	ctx := context.Background()
	gameUI, err := gameSvc.CreateGame(ctx, "test-game-eval", "owner-1")
	if err != nil {
		t.Fatalf("CreateGame failed: %v", err)
	}

	payload := []byte(`{"action":"evaluate","options":{"action_list":[["add",1,2],["sub",5,3],["plugin_card","p-1","card-1"],["obj_set_by_path",{"a":{"b":1}},["a","b"],42],["reduce",["list",1,2,3],0,["add"]],["rand",5]]}}`)
	if err := gameSvc.SendAction(gameUI.Slug, payload); err != nil {
		t.Fatalf("SendAction failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	roomActions, err := roomSvc.ListActions(gameUI.Slug)
	if err != nil {
		t.Fatalf("RoomService list actions: %v", err)
	}
	if len(roomActions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(roomActions))
	}
}
