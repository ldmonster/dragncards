package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ldmonster/dragncards/ha-be/internal/application/game"
	"github.com/ldmonster/dragncards/ha-be/internal/application/replay"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/lfg"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/persistence"
	"github.com/ldmonster/dragncards/ha-be/internal/platform/auth"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// drainUntil reads from c until an envelope with the target event is found.
// It skips server-push events like users_changed, seats_changed, spectators_changed, gui_update.
func drainUntil(t *testing.T, ctx context.Context, c *websocket.Conn, targetEvent string) PhoenixEnvelope {
	t.Helper()
	for {
		var env PhoenixEnvelope
		if err := wsjson.Read(ctx, c, &env); err != nil {
			t.Fatalf("read while draining for %s: %v", targetEvent, err)
		}
		if env.Event == targetEvent {
			return env
		}
		// allow server-push side-channel events to pass through silently
		switch env.Event {
		case "users_changed", "seats_changed", "spectators_changed", "gui_update", "current_state":
			continue
		default:
			t.Fatalf("unexpected event %q while waiting for %q", env.Event, targetEvent)
		}
	}
}

func TestRoomChannelGameActionBroadcast(t *testing.T) {
	hub := NewHub()
	rRepo := persistence.NewInMemoryRoomRepository()
	rSvc := room.NewService(rRepo)
	lSvc := lfg.NewService(persistence.NewInMemoryLfgRepository())
	gameRegistry := game.NewRoomRegistry()
	gameSvc := game.NewGameService(rSvc, gameRegistry, nil)
	gameUI, err := gameSvc.CreateGame(context.Background(), "test", "owner")
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	topic := "room:" + gameUI.Slug
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	handler := NewWSHandler(hub, rSvc, lSvc, gameSvc, replaySvc)
	server := httptest.NewServer(handler)
	defer server.Close()

	token1, err := auth.GenerateToken("userA", time.Hour)
	if err != nil {
		t.Fatalf("generate token1: %v", err)
	}
	wsURL1 := "ws" + strings.TrimPrefix(server.URL, "http") + "/be/socket?token=" + token1

	token2, err := auth.GenerateToken("userB", time.Hour)
	if err != nil {
		t.Fatalf("generate token2: %v", err)
	}
	wsURL2 := "ws" + strings.TrimPrefix(server.URL, "http") + "/be/socket?token=" + token2

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c1, _, err := websocket.Dial(ctx, wsURL1, nil)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	c2, _, err := websocket.Dial(ctx, wsURL2, nil)
	if err != nil {
		t.Fatalf("dial c2: %v", err)
	}
	defer c2.Close(websocket.StatusNormalClosure, "")

	join1 := PhoenixEnvelope{Topic: topic, Event: "phx_join", Payload: json.RawMessage(`{"when":"t1"}`)}
	if err := wsjson.Write(ctx, c1, join1); err != nil {
		t.Fatalf("c1 join write: %v", err)
	}
	_ = drainUntil(t, ctx, c1, "phx_reply")

	join2 := PhoenixEnvelope{Topic: topic, Event: "phx_join", Payload: json.RawMessage(`{"when":"t2"}`)}
	if err := wsjson.Write(ctx, c2, join2); err != nil {
		t.Fatalf("c2 join write: %v", err)
	}
	_ = drainUntil(t, ctx, c2, "phx_reply")

	action := PhoenixEnvelope{Topic: topic, Event: "game_action", Payload: json.RawMessage(`{"actor":"c2","action":"move"}`)}
	if err := wsjson.Write(ctx, c2, action); err != nil {
		t.Fatalf("c2 game_action write: %v", err)
	}

	update := drainUntil(t, ctx, c1, "send_update")
	if string(update.Payload) != `{"actor":"c2","action":"move"}` {
		t.Fatalf("unexpected payload: %s", string(update.Payload))
	}

	// verify persisted action in room service
	actions, err := rSvc.ListActions(gameUI.Slug)
	if err != nil {
		t.Fatalf("list actions: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 persisted action, got %d", len(actions))
	}
	if string(actions[0]) != `{"actor":"c2","action":"move"}` {
		t.Fatalf("unexpected persisted action: %s", string(actions[0]))
	}

	state, err := gameSvc.GetGameUI(gameUI.Slug)
	if err != nil {
		t.Fatalf("get game ui: %v", err)
	}
	if len(state.Actions) != 1 {
		t.Fatalf("expected game ui to have 1 action, got %d", len(state.Actions))
	}
}

func TestRoomChannelSetSeatSpectatorState(t *testing.T) {
	hub := NewHub()
	rRepo := persistence.NewInMemoryRoomRepository()
	rSvc := room.NewService(rRepo)
	lSvc := lfg.NewService(persistence.NewInMemoryLfgRepository())
	gameRegistry := game.NewRoomRegistry()
	gameSvc := game.NewGameService(rSvc, gameRegistry, nil)
	gameUI, err := gameSvc.CreateGame(context.Background(), "test", "owner")
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	topic := "room:" + gameUI.Slug
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	handler := NewWSHandler(hub, rSvc, lSvc, gameSvc, replaySvc)
	server := httptest.NewServer(handler)
	defer server.Close()

	token, err := auth.GenerateToken("userA", time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/be/socket?token=" + token

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c1, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	join := PhoenixEnvelope{Topic: topic, Event: "phx_join", Payload: json.RawMessage(`{"when":"t1"}`)}
	if err := wsjson.Write(ctx, c1, join); err != nil {
		t.Fatalf("c1 join write: %v", err)
	}
	_ = drainUntil(t, ctx, c1, "phx_reply")

	// set_seat via offline POST
	seatEnv := PhoenixEnvelope{Topic: topic, Event: "set_seat", Payload: json.RawMessage(`{"player_id":"userA","seat":"A1"}`)}
	seatBody, _ := json.Marshal(seatEnv)
	res, err := http.Post(server.URL+"/be/socket", "application/json", strings.NewReader(string(seatBody)))
	if err != nil {
		t.Fatalf("post set_seat failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}
	_ = drainUntil(t, ctx, c1, "set_seat")
	seatsChanged := drainUntil(t, ctx, c1, "seats_changed")
	var seatsResp struct {
		Seats map[string]string `json:"seats"`
	}
	if err := json.Unmarshal(seatsChanged.Payload, &seatsResp); err != nil {
		t.Fatalf("decode seats_changed payload: %v", err)
	}
	if seatsResp.Seats["userA"] != "A1" {
		t.Fatalf("expected userA seat A1, got %q", seatsResp.Seats["userA"])
	}

	// set_spectator via offline POST
	specEnv := PhoenixEnvelope{Topic: topic, Event: "set_spectator", Payload: json.RawMessage(`{"player_id":"userA","spectator":true}`)}
	specBody, _ := json.Marshal(specEnv)
	res, err = http.Post(server.URL+"/be/socket", "application/json", strings.NewReader(string(specBody)))
	if err != nil {
		t.Fatalf("post set_spectator failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}
	_ = drainUntil(t, ctx, c1, "set_spectator")
	specChanged := drainUntil(t, ctx, c1, "spectators_changed")
	var specResp struct {
		Spectators map[string]bool `json:"spectators"`
	}
	if err := json.Unmarshal(specChanged.Payload, &specResp); err != nil {
		t.Fatalf("decode spectators_changed payload: %v", err)
	}
	if !specResp.Spectators["userA"] {
		t.Fatalf("expected userA in spectators")
	}

	updated, err := gameSvc.GetGameUI(gameUI.Slug)
	if err != nil {
		t.Fatalf("get game ui: %v", err)
	}
	if updated.Seats["userA"] != "A1" {
		t.Fatalf("expected persisted seat userA=A1, got %q", updated.Seats["userA"])
	}
	if !updated.Spectators["userA"] {
		t.Fatalf("expected persisted spectator userA")
	}
}

func TestLfgChannelNewPostPersistsAndBroadcasts(t *testing.T) {
	hub := NewHub()
	lfgSvc := lfg.NewService(persistence.NewInMemoryLfgRepository())
	rSvc := room.NewService(persistence.NewInMemoryRoomRepository())
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	handler := NewWSHandler(hub, rSvc, lfgSvc, nil, replaySvc)
	server := httptest.NewServer(handler)
	defer server.Close()

	token, err := auth.GenerateToken("userA", time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/be/socket?token=" + token

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c1, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	// subscribe c1 to plugin channel
	join := PhoenixEnvelope{Topic: "lfg:plugin123", Event: "phx_join", Payload: json.RawMessage(`{"when":"t1"}`)}
	if err := wsjson.Write(ctx, c1, join); err != nil {
		t.Fatalf("c1 join write: %v", err)
	}
	var resp PhoenixEnvelope
	if err := wsjson.Read(ctx, c1, &resp); err != nil {
		t.Fatalf("c1 join read: %v", err)
	}

	// post new LFG message via same socket
	newPost := PhoenixEnvelope{Topic: "lfg:plugin123", Event: "new_post", Payload: json.RawMessage(`{"plugin_id":"plugin123","user_id":"userA","text":"Looking for game"}`)}
	if err := wsjson.Write(ctx, c1, newPost); err != nil {
		t.Fatalf("c1 new_post write: %v", err)
	}

	var broadcast PhoenixEnvelope
	if err := wsjson.Read(ctx, c1, &broadcast); err != nil {
		t.Fatalf("c1 broadcast read: %v", err)
	}
	if broadcast.Event != "new_post" {
		t.Fatalf("expected new_post broadcast, got %s", broadcast.Event)
	}

	posts, err := lfgSvc.List("plugin123")
	if err != nil {
		t.Fatalf("lfg list: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 persisted lfg post, got %d", len(posts))
	}
	if posts[0].Text != "Looking for game" {
		t.Fatalf("unexpected lfg text: %s", posts[0].Text)
	}
}

func TestWSHandlerPostOfflineClientIDUnique(t *testing.T) {
	hub := NewHub()
	rSvc := room.NewService(persistence.NewInMemoryRoomRepository())
	lSvc := lfg.NewService(persistence.NewInMemoryLfgRepository())
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	handler := NewWSHandler(hub, rSvc, lSvc, nil, replaySvc)
	server := httptest.NewServer(handler)
	defer server.Close()

	postClientID := func() string {
		env := PhoenixEnvelope{Topic: "room:test", Event: "phx_join", Payload: json.RawMessage(`{"when":"t"}`)}
		body, _ := json.Marshal(env)
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/be/socket", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("post request failed: %v", err)
		}
		defer res.Body.Close()
		var out struct {
			Status   string `json:"status"`
			ClientID string `json:"client_id"`
		}
		if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if out.Status != "joined" {
			t.Fatalf("expected joined status, got %q", out.Status)
		}
		if out.ClientID == "" {
			t.Fatalf("expected client_id in response")
		}
		return out.ClientID
	}

	id1 := postClientID()
	id2 := postClientID()
	if id1 == id2 {
		t.Fatalf("expected unique client ids, got %q and %q", id1, id2)
	}
}

func TestWSHandlerPostOfflineGameActionBroadcastAndPersist(t *testing.T) {
	hub := NewHub()
	rRepo := persistence.NewInMemoryRoomRepository()
	rSvc := room.NewService(rRepo)
	lSvc := lfg.NewService(persistence.NewInMemoryLfgRepository())
	roomObj, err := rSvc.Create("test", "owner")
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	topic := "room:" + roomObj.Slug
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	handler := NewWSHandler(hub, rSvc, lSvc, nil, replaySvc)
	server := httptest.NewServer(handler)
	defer server.Close()

	token, err := auth.GenerateToken("userA", time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/be/socket?token=" + token

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c1, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	join := PhoenixEnvelope{Topic: topic, Event: "phx_join", Payload: json.RawMessage(`{"when":"t1"}`)}
	if err := wsjson.Write(ctx, c1, join); err != nil {
		t.Fatalf("c1 join write: %v", err)
	}
	_ = drainUntil(t, ctx, c1, "phx_reply")

	// Issue offline POST game_action
	env := PhoenixEnvelope{Topic: topic, Event: "game_action", Payload: json.RawMessage(`{"actor":"offline","action":"move"}`)}
	body, _ := json.Marshal(env)
	res, err := http.Post(server.URL+"/be/socket", "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("post game_action failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}

	update := drainUntil(t, ctx, c1, "send_update")
	if string(update.Payload) != `{"actor":"offline","action":"move"}` {
		t.Fatalf("unexpected payload: %s", string(update.Payload))
	}

	actions, err := rSvc.ListActions(roomObj.Slug)
	if err != nil {
		t.Fatalf("list actions: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 persisted action, got %d", len(actions))
	}
	if string(actions[0]) != `{"actor":"offline","action":"move"}` {
		t.Fatalf("unexpected persisted action: %s", string(actions[0]))
	}
}

func TestWSHandlerPostOfflineRequestState(t *testing.T) {
	hub := NewHub()
	rRepo := persistence.NewInMemoryRoomRepository()
	rSvc := room.NewService(rRepo)
	lSvc := lfg.NewService(persistence.NewInMemoryLfgRepository())
	roomObj, err := rSvc.Create("test", "owner")
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	topic := "room:" + roomObj.Slug
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	handler := NewWSHandler(hub, rSvc, lSvc, nil, replaySvc)
	server := httptest.NewServer(handler)
	defer server.Close()

	token, err := auth.GenerateToken("userA", time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/be/socket?token=" + token

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c1, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	join := PhoenixEnvelope{Topic: topic, Event: "phx_join", Payload: json.RawMessage(`{"when":"t1"}`)}
	if err := wsjson.Write(ctx, c1, join); err != nil {
		t.Fatalf("c1 join write: %v", err)
	}
	_ = drainUntil(t, ctx, c1, "phx_reply")

	action := PhoenixEnvelope{Topic: topic, Event: "game_action", Payload: json.RawMessage(`{"actor":"offline","action":"move"}`)}
	if err := wsjson.Write(ctx, c1, action); err != nil {
		t.Fatalf("c1 game_action write: %v", err)
	}
	_ = drainUntil(t, ctx, c1, "send_update")

	requestState := PhoenixEnvelope{Topic: topic, Event: "request_state"}
	body, _ := json.Marshal(requestState)
	res, err := http.Post(server.URL+"/be/socket", "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("post request_state failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}
	var stateResp PhoenixEnvelope
	if err := json.NewDecoder(res.Body).Decode(&stateResp); err != nil {
		t.Fatalf("decode state response: %v", err)
	}
	if stateResp.Event != "current_state" {
		t.Fatalf("expected current_state response, got %s", stateResp.Event)
	}
	var payload struct {
		Actions [][]byte `json:"actions"`
	}
	if err := json.Unmarshal(stateResp.Payload, &payload); err != nil {
		t.Fatalf("decode payload actions: %v", err)
	}
	if len(payload.Actions) != 1 {
		t.Fatalf("expected 1 action in state, got %d", len(payload.Actions))
	}
	if string(payload.Actions[0]) != `{"actor":"offline","action":"move"}` {
		t.Fatalf("unexpected state action: %s", string(payload.Actions[0]))
	}
}

func TestRoomChannelPhxJoinCurrentState(t *testing.T) {
	hub := NewHub()
	rRepo := persistence.NewInMemoryRoomRepository()
	rSvc := room.NewService(rRepo)
	gameRegistry := game.NewRoomRegistry()
	gameSvc := game.NewGameService(rSvc, gameRegistry, nil)
	gameUI, err := gameSvc.CreateGame(context.Background(), "test", "owner")
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	topic := "room:" + gameUI.Slug
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	handler := NewWSHandler(hub, rSvc, lfg.NewService(persistence.NewInMemoryLfgRepository()), gameSvc, replaySvc)
	server := httptest.NewServer(handler)
	defer server.Close()

	token, err := auth.GenerateToken("userA", time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/be/socket?token=" + token

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c1, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	join := PhoenixEnvelope{Topic: topic, Event: "phx_join", Payload: json.RawMessage(`{"when":"t1"}`)}
	if err := wsjson.Write(ctx, c1, join); err != nil {
		t.Fatalf("c1 join write: %v", err)
	}
	_ = drainUntil(t, ctx, c1, "phx_reply")

	state := drainUntil(t, ctx, c1, "current_state")
	var statePayload struct {
		Actions [][]byte `json:"actions"`
	}
	if err := json.Unmarshal(state.Payload, &statePayload); err != nil {
		t.Fatalf("decode current_state payload: %v", err)
	}
	if len(statePayload.Actions) != 0 {
		t.Fatalf("expected 0 actions in current_state, got %d", len(statePayload.Actions))
	}
}

func TestWSHandlerPostOfflineResetGameBroadcast(t *testing.T) {
	hub := NewHub()
	rRepo := persistence.NewInMemoryRoomRepository()
	rSvc := room.NewService(rRepo)
	lSvc := lfg.NewService(persistence.NewInMemoryLfgRepository())
	roomObj, err := rSvc.Create("test", "owner")
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	topic := "room:" + roomObj.Slug
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	handler := NewWSHandler(hub, rSvc, lSvc, nil, replaySvc)
	server := httptest.NewServer(handler)
	defer server.Close()

	token, err := auth.GenerateToken("userA", time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/be/socket?token=" + token

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c1, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	join := PhoenixEnvelope{Topic: topic, Event: "phx_join", Payload: json.RawMessage(`{"when":"t1"}`)}
	if err := wsjson.Write(ctx, c1, join); err != nil {
		t.Fatalf("c1 join write: %v", err)
	}
	_ = drainUntil(t, ctx, c1, "phx_reply")

	reset := PhoenixEnvelope{Topic: topic, Event: "reset_game", Payload: json.RawMessage(`{"reason":"test-reset"}`)}
	body, _ := json.Marshal(reset)
	res, err := http.Post(server.URL+"/be/socket", "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("post reset_game failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}
	var pr PhoenixEnvelope
	if err := json.NewDecoder(res.Body).Decode(&pr); err != nil {
		t.Fatalf("decode reset response: %v", err)
	}
	if pr.Event != "phx_reply" {
		t.Fatalf("expected phx_reply response, got %s", pr.Event)
	}
	broadcast := drainUntil(t, ctx, c1, "reset_game")
	_ = broadcast
}
