package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ldmonster/dragncards/ha-be/internal/application/deck"
	"github.com/ldmonster/dragncards/ha-be/internal/application/game"
	"github.com/ldmonster/dragncards/ha-be/internal/application/replay"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/alert"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/identity"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/lfg"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/plugin"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/persistence"
	httpapi "github.com/ldmonster/dragncards/ha-be/internal/interfaces/http"
	"github.com/ldmonster/dragncards/ha-be/internal/platform/auth"
)

func TestHealthEndpoint(t *testing.T) {
	userRepo := persistence.NewInMemoryUserRepository()
	identitySvc := identity.NewService(userRepo)

	roomRepo := persistence.NewInMemoryRoomRepository()
	roomSvc := room.NewService(roomRepo)

	pluginRepo := persistence.NewInMemoryPluginRepository()
	pluginSvc := plugin.NewService(pluginRepo)

	lfgRepo := persistence.NewInMemoryLfgRepository()
	lfgSvc := lfg.NewService(lfgRepo)

	alertRepo := persistence.NewInMemoryAlertRepository()
	alertSvc := alert.NewService(alertRepo)

	deckSvc := deck.NewDeckService(persistence.NewInMemoryDeckRepository())
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	gameRegistry := game.NewRoomRegistry()
	gameSvc := game.NewGameService(roomSvc, gameRegistry, nil)
	apiHandler := httpapi.NewAPIHandlerLegacy(identitySvc, roomSvc, pluginSvc, gameSvc, deckSvc, replaySvc, lfgSvc, alertSvc)
	mux := httpapi.NewRouter(apiHandler)

	req := httptest.NewRequest(http.MethodGet, "/be/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	userRepo := persistence.NewInMemoryUserRepository()
	identitySvc := identity.NewService(userRepo)

	roomRepo := persistence.NewInMemoryRoomRepository()
	roomSvc := room.NewService(roomRepo)

	pluginRepo := persistence.NewInMemoryPluginRepository()
	pluginSvc := plugin.NewService(pluginRepo)

	lfgRepo := persistence.NewInMemoryLfgRepository()
	lfgSvc := lfg.NewService(lfgRepo)

	alertRepo := persistence.NewInMemoryAlertRepository()
	alertSvc := alert.NewService(alertRepo)

	deckSvc := deck.NewDeckService(persistence.NewInMemoryDeckRepository())
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	gameRegistry := game.NewRoomRegistry()
	gameSvc := game.NewGameService(roomSvc, gameRegistry, nil)
	apiHandler := httpapi.NewAPIHandlerLegacy(identitySvc, roomSvc, pluginSvc, gameSvc, deckSvc, replaySvc, lfgSvc, alertSvc)
	mux := httpapi.NewRouter(apiHandler)

	regBody := `{"email":"test@x.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/be/api/v1/registration", strings.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 got %d", w.Result().StatusCode)
	}

	loginBody := `{"email":"test@x.com","password":"pass"}`
	req = httptest.NewRequest(http.MethodPost, "/be/api/v1/session", strings.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Result().StatusCode)
	}

	var out map[string]string
	if err := json.NewDecoder(w.Result().Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out["auth_token"] == "" || out["renew_token"] == "" {
		t.Fatal("tokens missing")
	}

	// renew token
	renewBody := `{"renew_token":"` + out["renew_token"] + `"}`
	req = httptest.NewRequest(http.MethodPost, "/be/api/v1/session/renew", strings.NewReader(renewBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Result().StatusCode)
	}

	var renewResp map[string]string
	if err := json.NewDecoder(w.Result().Body).Decode(&renewResp); err != nil {
		t.Fatalf("decode renew response: %v", err)
	}
	if renewResp["auth_token"] == "" || renewResp["renew_token"] == "" {
		t.Fatal("renew tokens missing")
	}
}

func TestLogoutAndTokenRevoke(t *testing.T) {
	userRepo := persistence.NewInMemoryUserRepository()
	identitySvc := identity.NewService(userRepo)

	roomRepo := persistence.NewInMemoryRoomRepository()
	roomSvc := room.NewService(roomRepo)

	pluginRepo := persistence.NewInMemoryPluginRepository()
	pluginSvc := plugin.NewService(pluginRepo)

	lfgRepo := persistence.NewInMemoryLfgRepository()
	lfgSvc := lfg.NewService(lfgRepo)

	alertRepo := persistence.NewInMemoryAlertRepository()
	alertSvc := alert.NewService(alertRepo)

	deckSvc := deck.NewDeckService(persistence.NewInMemoryDeckRepository())
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	gameRegistry := game.NewRoomRegistry()
	gameSvc := game.NewGameService(roomSvc, gameRegistry, nil)
	apiHandler := httpapi.NewAPIHandlerLegacy(identitySvc, roomSvc, pluginSvc, gameSvc, deckSvc, replaySvc, lfgSvc, alertSvc)
	mux := httpapi.NewRouter(apiHandler)

	regBody := `{"email":"logout@x.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/be/api/v1/registration", strings.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 got %d", w.Result().StatusCode)
	}

	loginBody := `{"email":"logout@x.com","password":"pass"}`
	req = httptest.NewRequest(http.MethodPost, "/be/api/v1/session", strings.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Result().StatusCode)
	}

	var out map[string]string
	if err := json.NewDecoder(w.Result().Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out["auth_token"] == "" {
		t.Fatal("auth token missing")
	}

	logoutBody := `{"auth_token":"` + out["auth_token"] + `"}`
	req = httptest.NewRequest(http.MethodDelete, "/be/api/v1/session", strings.NewReader(logoutBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 got %d", w.Result().StatusCode)
	}

	// revoked token should be rejected by ParseToken
	if _, err := auth.ParseToken(out["auth_token"]); err == nil {
		t.Fatal("expected revoked token to be invalid")
	}
}

func TestEmailConfirmAndReset(t *testing.T) {
	userRepo := persistence.NewInMemoryUserRepository()
	identitySvc := identity.NewService(userRepo)

	roomRepo := persistence.NewInMemoryRoomRepository()
	roomSvc := room.NewService(roomRepo)

	pluginRepo := persistence.NewInMemoryPluginRepository()
	pluginSvc := plugin.NewService(pluginRepo)

	lfgRepo := persistence.NewInMemoryLfgRepository()
	lfgSvc := lfg.NewService(lfgRepo)

	alertRepo := persistence.NewInMemoryAlertRepository()
	alertSvc := alert.NewService(alertRepo)

	deckSvc := deck.NewDeckService(persistence.NewInMemoryDeckRepository())
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	gameRegistry := game.NewRoomRegistry()
	gameSvc := game.NewGameService(roomSvc, gameRegistry, nil)
	apiHandler := httpapi.NewAPIHandlerLegacy(identitySvc, roomSvc, pluginSvc, gameSvc, deckSvc, replaySvc, lfgSvc, alertSvc)
	mux := httpapi.NewRouter(apiHandler)

	regBody := `{"email":"abc@x.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/be/api/v1/registration", strings.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusCreated {
		t.Fatalf("registration expected 201 got %d", w.Result().StatusCode)
	}

	// confirm email first
	confirmToken, err := identitySvc.GenerateConfirmToken("abc@x.com")
	if err != nil {
		t.Fatalf("generate confirm token: %v", err)
	}
	if err := identitySvc.ConfirmEmail(confirmToken); err != nil {
		t.Fatalf("confirm email: %v", err)
	}

	// request reset
	resetReq := `{"email":"abc@x.com"}`
	req = httptest.NewRequest(http.MethodPost, "/be/api/v1/reset-password", strings.NewReader(resetReq))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("reset request expected 200 got %d", w.Result().StatusCode)
	}
	var resetResp map[string]string
	if err := json.NewDecoder(w.Result().Body).Decode(&resetResp); err != nil {
		t.Fatalf("decode reset resp: %v", err)
	}
	resetToken := resetResp["reset_token"]
	if resetToken == "" {
		t.Fatal("reset token missing")
	}

	updateReq := `{"token":"` + resetToken + `","password":"newpass"}`
	req = httptest.NewRequest(http.MethodPost, "/be/api/v1/reset-password/update", strings.NewReader(updateReq))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusNoContent {
		t.Fatalf("reset update expected 204 got %d", w.Result().StatusCode)
	}

	// ability to login with new password
	loginReq := `{"email":"abc@x.com","password":"newpass"}`
	req = httptest.NewRequest(http.MethodPost, "/be/api/v1/session", strings.NewReader(loginReq))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("login after reset expected 200 got %d", w.Result().StatusCode)
	}
}

func TestRoomsAndPluginsEndpoints(t *testing.T) {
	userRepo := persistence.NewInMemoryUserRepository()
	identitySvc := identity.NewService(userRepo)

	roomRepo := persistence.NewInMemoryRoomRepository()
	roomSvc := room.NewService(roomRepo)

	pluginRepo := persistence.NewInMemoryPluginRepository()
	pluginSvc := plugin.NewService(pluginRepo)

	lfgRepo := persistence.NewInMemoryLfgRepository()
	lfgSvc := lfg.NewService(lfgRepo)

	alertRepo := persistence.NewInMemoryAlertRepository()
	alertSvc := alert.NewService(alertRepo)

	deckSvc := deck.NewDeckService(persistence.NewInMemoryDeckRepository())
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	gameRegistry := game.NewRoomRegistry()
	gameSvc := game.NewGameService(roomSvc, gameRegistry, nil)
	apiHandler := httpapi.NewAPIHandlerLegacy(identitySvc, roomSvc, pluginSvc, gameSvc, deckSvc, replaySvc, lfgSvc, alertSvc)
	mux := httpapi.NewRouter(apiHandler)

	// room create
	roomBody := `{"name":"test room","owner_id":"u-1"}`
	req := httptest.NewRequest(http.MethodPost, "/be/api/rooms", strings.NewReader(roomBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 got %d", w.Result().StatusCode)
	}

	// room list
	req = httptest.NewRequest(http.MethodGet, "/be/api/rooms", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Result().StatusCode)
	}

	var rooms []room.Room
	if err := json.NewDecoder(w.Result().Body).Decode(&rooms); err != nil {
		t.Fatalf("decode rooms: %v", err)
	}
	if len(rooms) != 1 {
		t.Fatalf("expected 1 room got %d", len(rooms))
	}

	// plugin create and list
	if _, err := pluginSvc.Create("core", true); err != nil {
		t.Fatalf("create plugin: %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/be/api/plugins", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Result().StatusCode)
	}
	var plugins []plugin.Plugin
	if err := json.NewDecoder(w.Result().Body).Decode(&plugins); err != nil {
		t.Fatalf("decode plugins: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin got %d", len(plugins))
	}

	req = httptest.NewRequest(http.MethodGet, "/be/api/plugins/visible", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Result().StatusCode)
	}
	var visible []plugin.Plugin
	if err := json.NewDecoder(w.Result().Body).Decode(&visible); err != nil {
		t.Fatalf("decode visible plugins: %v", err)
	}
	if len(visible) != 1 {
		t.Fatalf("expected 1 visible plugin got %d", len(visible))
	}
}

func TestLfgAndAlertsRequireAuth(t *testing.T) {
	userRepo := persistence.NewInMemoryUserRepository()
	identitySvc := identity.NewService(userRepo)

	roomRepo := persistence.NewInMemoryRoomRepository()
	roomSvc := room.NewService(roomRepo)

	pluginRepo := persistence.NewInMemoryPluginRepository()
	pluginSvc := plugin.NewService(pluginRepo)

	lfgRepo := persistence.NewInMemoryLfgRepository()
	lfgSvc := lfg.NewService(lfgRepo)

	alertRepo := persistence.NewInMemoryAlertRepository()
	alertSvc := alert.NewService(alertRepo)

	deckSvc := deck.NewDeckService(persistence.NewInMemoryDeckRepository())
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	gameRegistry := game.NewRoomRegistry()
	gameSvc := game.NewGameService(roomSvc, gameRegistry, nil)
	apiHandler := httpapi.NewAPIHandlerLegacy(identitySvc, roomSvc, pluginSvc, gameSvc, deckSvc, replaySvc, lfgSvc, alertSvc)
	mux := httpapi.NewRouter(apiHandler)

	// register + login
	regBody := `{"email":"auth@x.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/be/api/v1/registration", strings.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 got %d", w.Result().StatusCode)
	}

	loginBody := `{"email":"auth@x.com","password":"pass"}`
	req = httptest.NewRequest(http.MethodPost, "/be/api/v1/session", strings.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Result().StatusCode)
	}
	var out map[string]string
	if err := json.NewDecoder(w.Result().Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	authToken := out["auth_token"]
	if authToken == "" {
		t.Fatal("missing auth token")
	}

	// without auth header
	req = httptest.NewRequest(http.MethodGet, "/be/api/lfg", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", w.Result().StatusCode)
	}

	// invalid token
	req = httptest.NewRequest(http.MethodGet, "/be/api/lfg", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", w.Result().StatusCode)
	}

	// valid token
	req = httptest.NewRequest(http.MethodGet, "/be/api/lfg", nil)
	req.Header.Set("Authorization", "Bearer "+authToken)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Result().StatusCode)
	}
	var listings []lfg.LfgPost
	if err := json.NewDecoder(w.Result().Body).Decode(&listings); err != nil {
		t.Fatalf("decode lfg listings: %v", err)
	}
	if len(listings) != 0 {
		t.Fatalf("expected 0 listings got %d", len(listings))
	}
}

func TestPluginCreateAndCustomCardRoutes(t *testing.T) {
	userRepo := persistence.NewInMemoryUserRepository()
	identitySvc := identity.NewService(userRepo)

	roomRepo := persistence.NewInMemoryRoomRepository()
	roomSvc := room.NewService(roomRepo)

	pluginRepo := persistence.NewInMemoryPluginRepository()
	pluginSvc := plugin.NewService(pluginRepo)

	lfgRepo := persistence.NewInMemoryLfgRepository()
	lfgSvc := lfg.NewService(lfgRepo)

	alertRepo := persistence.NewInMemoryAlertRepository()
	alertSvc := alert.NewService(alertRepo)

	deckSvc := deck.NewDeckService(persistence.NewInMemoryDeckRepository())
	replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
	gameRegistry := game.NewRoomRegistry()
	gameSvc := game.NewGameService(roomSvc, gameRegistry, nil)
	apiHandler := httpapi.NewAPIHandlerLegacy(identitySvc, roomSvc, pluginSvc, gameSvc, deckSvc, replaySvc, lfgSvc, alertSvc)
	mux := httpapi.NewRouter(apiHandler)

	pluginBody := `{"name":"test-plugin","visible":true}`
	req := httptest.NewRequest(http.MethodPost, "/be/api/plugins", strings.NewReader(pluginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on plugin create got %d", w.Result().StatusCode)
	}

	var createdPlugin plugin.Plugin
	if err := json.NewDecoder(w.Result().Body).Decode(&createdPlugin); err != nil {
		t.Fatalf("decode plugin create resp: %v", err)
	}

	req = httptest.NewRequest(http.MethodGet, "/be/api/plugins/visible", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on visible plugins got %d", w.Result().StatusCode)
	}

	var visible []plugin.Plugin
	if err := json.NewDecoder(w.Result().Body).Decode(&visible); err != nil {
		t.Fatalf("decode visible plugins: %v", err)
	}
	if len(visible) != 1 || visible[0].ID != createdPlugin.ID {
		t.Fatalf("expected one visible plugin, got %v", visible)
	}

	customCardBody := `{"name":"test-card","data":"{\"foo\":\"bar\"}"}`
	url := "/be/api/plugins/" + createdPlugin.ID + "/cards"
	req = httptest.NewRequest(http.MethodPost, url, strings.NewReader(customCardBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on custom card create got %d", w.Result().StatusCode)
	}
	var outCard plugin.CustomCard
	if err := json.NewDecoder(w.Result().Body).Decode(&outCard); err != nil {
		t.Fatalf("decode create custom card: %v", err)
	}
	if outCard.PluginID != createdPlugin.ID {
		t.Fatalf("custom card plugin id mismatch: %s", outCard.PluginID)
	}

	req = httptest.NewRequest(http.MethodGet, url, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on list custom cards got %d", w.Result().StatusCode)
	}
	var cardList []plugin.CustomCard
	if err := json.NewDecoder(w.Result().Body).Decode(&cardList); err != nil {
		t.Fatalf("decode list custom cards: %v", err)
	}
	if len(cardList) != 1 || cardList[0].Name != "test-card" {
		t.Fatalf("unexpected custom cards list: %v", cardList)
	}
}

func TestAdminAndPluginPermissionEndpoints(t *testing.T) {
userRepo := persistence.NewInMemoryUserRepository()
identitySvc := identity.NewService(userRepo)

roomRepo := persistence.NewInMemoryRoomRepository()
roomSvc := room.NewService(roomRepo)

pluginRepo := persistence.NewInMemoryPluginRepository()
pluginSvc := plugin.NewService(pluginRepo)

lfgRepo := persistence.NewInMemoryLfgRepository()
lfgSvc := lfg.NewService(lfgRepo)

alertRepo := persistence.NewInMemoryAlertRepository()
alertSvc := alert.NewService(alertRepo)

deckSvc := deck.NewDeckService(persistence.NewInMemoryDeckRepository())
replaySvc := replay.NewReplayService(persistence.NewInMemoryReplayRepository())
gameRegistry := game.NewRoomRegistry()
gameSvc := game.NewGameService(roomSvc, gameRegistry, nil)
apiHandler := httpapi.NewAPIHandlerLegacy(identitySvc, roomSvc, pluginSvc, gameSvc, deckSvc, replaySvc, lfgSvc, alertSvc)
mux := httpapi.NewRouter(apiHandler)

// register user
regBody := `{"email":"admin@x.com","password":"pass"}`
req := httptest.NewRequest(http.MethodPost, "/be/api/v1/registration", strings.NewReader(regBody))
req.Header.Set("Content-Type", "application/json")
w := httptest.NewRecorder()
mux.ServeHTTP(w, req)
if w.Result().StatusCode != http.StatusCreated {
t.Fatalf("expected 201 got %d", w.Result().StatusCode)
}

// login
loginBody := `{"email":"admin@x.com","password":"pass"}`
req = httptest.NewRequest(http.MethodPost, "/be/api/v1/session", strings.NewReader(loginBody))
req.Header.Set("Content-Type", "application/json")
w = httptest.NewRecorder()
mux.ServeHTTP(w, req)
if w.Result().StatusCode != http.StatusOK {
t.Fatalf("expected 200 got %d", w.Result().StatusCode)
}
var out map[string]string
if err := json.NewDecoder(w.Result().Body).Decode(&out); err != nil {
t.Fatalf("decode response: %v", err)
}
if out["auth_token"] == "" {
t.Fatal("auth token missing")
}

// plugin permission check
req = httptest.NewRequest(http.MethodGet, "/be/api/v1/users/plugin_permission/plugin123/user123", nil)
req.Header.Set("Authorization", "Bearer "+out["auth_token"])
w = httptest.NewRecorder()
mux.ServeHTTP(w, req)
if w.Result().StatusCode != http.StatusOK {
t.Fatalf("expected 200 got %d", w.Result().StatusCode)
}

// admin contact
contactBody := `{"email":"admin@x.com","message":"help"}`
req = httptest.NewRequest(http.MethodPost, "/be/api/v1/admin_contact", strings.NewReader(contactBody))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+out["auth_token"])
w = httptest.NewRecorder()
mux.ServeHTTP(w, req)
if w.Result().StatusCode != http.StatusAccepted {
t.Fatalf("expected 202 got %d", w.Result().StatusCode)
}

// update patreon
patreonBody := `{"user_id":"user123","tier":"gold"}`
req = httptest.NewRequest(http.MethodPost, "/be/api/v1/admin/update_user_patreon", strings.NewReader(patreonBody))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+out["auth_token"])
w = httptest.NewRecorder()
mux.ServeHTTP(w, req)
if w.Result().StatusCode != http.StatusOK {
t.Fatalf("expected 200 got %d", w.Result().StatusCode)
}
}
