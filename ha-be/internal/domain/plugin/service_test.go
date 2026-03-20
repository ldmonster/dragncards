package plugin_test

import (
	"testing"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/plugin"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/persistence"
)

func TestPluginServiceCustomCards(t *testing.T) {
	repo := persistence.NewInMemoryPluginRepository()
	service := plugin.NewService(repo)

	plug, err := service.Create("TestPlugin", true)
	if err != nil {
		t.Fatalf("Create plugin failed: %v", err)
	}

	card, err := service.CreateCustomCard(plug.ID, "TestCard", `{"title":"X"}`)
	if err != nil {
		t.Fatalf("CreateCustomCard failed: %v", err)
	}
	if card.PluginID != plug.ID {
		t.Fatalf("wrong PluginID: %s", card.PluginID)
	}

	cards, err := service.ListCustomCards(plug.ID)
	if err != nil {
		t.Fatalf("ListCustomCards failed: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Data != `{"title":"X"}` {
		t.Fatalf("unexpected card data: %s", cards[0].Data)
	}
}
