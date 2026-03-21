package plugin_test

import (
	"net/http"
	"net/http/httptest"
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

	card, err := service.CreateCustomCard(plug.ID, "u-1", "TestCard", `{"title":"X"}`)
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

	// verify list by owner
	ownerCards, err := service.ListCustomCardsByOwner("u-1", plug.ID)
	if err != nil {
		t.Fatalf("ListCustomCardsByOwner failed: %v", err)
	}
	if len(ownerCards) != 1 {
		t.Fatalf("expected 1 owner card, got %d", len(ownerCards))
	}
	// delete custom card
	if err := service.DeleteCustomCard(card.ID); err != nil {
		t.Fatalf("DeleteCustomCard failed: %v", err)
	}
	cardsAfterDelete, err := service.ListCustomCards(plug.ID)
	if err != nil {
		t.Fatalf("ListCustomCards failed: %v", err)
	}
	if len(cardsAfterDelete) != 0 {
		t.Fatalf("expected 0 cards after delete, got %d", len(cardsAfterDelete))
	}
}

func TestPluginServiceSyncRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"plugins":[{"id":"p-sync","name":"SyncPlugin","visible":true,"cards":[{"id":"cc-sync","name":"SyncCard","data":"{\"foo\":\"bar\"}"}]},{"id":"p-no-cards","name":"NoCards","visible":false}]}`))
	}))
	defer server.Close()

	repo := persistence.NewInMemoryPluginRepository()
	service := plugin.NewService(repo)

	plugins, err := service.SyncRepository(server.URL)
	if err != nil {
		t.Fatalf("SyncRepository failed: %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("expected 2 synced plugins got %d", len(plugins))
	}

	p, err := service.FindByID("p-sync")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if p.Name != "SyncPlugin" || !p.Visible {
		t.Fatalf("sync plugin wrong: %+v", p)
	}

	cards, err := service.ListCustomCards("p-sync")
	if err != nil {
		t.Fatalf("ListCustomCards failed: %v", err)
	}
	if len(cards) != 1 || cards[0].ID != "cc-sync" {
		t.Fatalf("expected synced custom card got %+v", cards)
	}
}
