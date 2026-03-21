package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/ldmonster/dragncards/ha-be/internal/application/deck"
	"github.com/ldmonster/dragncards/ha-be/internal/application/game"
	identityapp "github.com/ldmonster/dragncards/ha-be/internal/application/identity"
	"github.com/ldmonster/dragncards/ha-be/internal/application/replay"
	settingsapp "github.com/ldmonster/dragncards/ha-be/internal/application/settings"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/alert"
	deckDomain "github.com/ldmonster/dragncards/ha-be/internal/domain/deck"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/identity"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/lfg"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/plugin"
	replayDomain "github.com/ldmonster/dragncards/ha-be/internal/domain/replay"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/settings"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/email"
	"github.com/ldmonster/dragncards/ha-be/internal/platform/auth"
	"github.com/ldmonster/dragncards/ha-be/internal/platform/middleware"
)

type APIHandler struct {
	identitySvc     *identityapp.Service
	roomSvc         *room.RoomService
	pluginSvc       *plugin.PluginService
	gameSvc         *game.GameService
	deckSvc         *deck.DeckService
	replaySvc       *replay.ReplayService
	lfgSvc          *lfg.LfgService
	alertSvc        *alert.AlertService
	settingsSvc     *settingsapp.Service
	mailer          email.Mailer
	recaptchaSecret string
	authTTL         time.Duration
	renewTTL        time.Duration
}

func NewAPIHandler(identitySvc *identityapp.Service, roomSvc *room.RoomService, pluginSvc *plugin.PluginService, gameSvc *game.GameService, deckSvc *deck.DeckService, replaySvc *replay.ReplayService, lfgSvc *lfg.LfgService, alertSvc *alert.AlertService, settingsSvc *settingsapp.Service, mailer email.Mailer, recaptchaSecret string, authTTL, renewTTL time.Duration) *APIHandler {
	return &APIHandler{
		identitySvc:     identitySvc,
		roomSvc:         roomSvc,
		pluginSvc:       pluginSvc,
		gameSvc:         gameSvc,
		deckSvc:         deckSvc,
		replaySvc:       replaySvc,
		lfgSvc:          lfgSvc,
		alertSvc:        alertSvc,
		settingsSvc:     settingsSvc,
		mailer:          mailer,
		recaptchaSecret: recaptchaSecret,
		authTTL:         authTTL,
		renewTTL:        renewTTL,
	}
}

func NewAPIHandlerLegacy(identitySvc *identityapp.Service, roomSvc *room.RoomService, pluginSvc *plugin.PluginService, gameSvc *game.GameService, deckSvc *deck.DeckService, replaySvc *replay.ReplayService, lfgSvc *lfg.LfgService, alertSvc *alert.AlertService) *APIHandler {
	return NewAPIHandler(identitySvc, roomSvc, pluginSvc, gameSvc, deckSvc, replaySvc, lfgSvc, alertSvc, nil, nil, "", 30*time.Minute, 90*24*time.Hour)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}

func (h *APIHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *APIHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	user, err := h.identitySvc.Register(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, identity.ErrUserExists) {
			writeError(w, http.StatusConflict, "user exists")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.identitySvc.GenerateConfirmToken(user.Email)
	if err == nil && h.mailer != nil {
		_ = h.mailer.SendConfirmation(user.Email, token)
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": user.ID, "email": user.Email, "confirmation_token": token})
}

func (h *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	user, err := h.identitySvc.Authenticate(req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tokenTTL := h.authTTL
	if tokenTTL <= 0 {
		tokenTTL = 30 * time.Minute
	}
	renewTTL := h.renewTTL
	if renewTTL <= 0 {
		renewTTL = 90 * 24 * time.Hour
	}
	authToken, _ := auth.GenerateToken(user.ID, tokenTTL)
	renewToken, _ := auth.GenerateToken(user.ID, renewTTL)
	writeJSON(w, http.StatusOK, map[string]any{"auth_token": authToken, "renew_token": renewToken})
}

func (h *APIHandler) RenewSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		RenewToken string `json:"renew_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.RenewToken == "" {
		writeError(w, http.StatusBadRequest, "renew_token required")
		return
	}

	userID, err := auth.ParseToken(req.RenewToken)
	if err != nil || userID == "" {
		writeError(w, http.StatusUnauthorized, "invalid renew token")
		return
	}

	tokenTTL := h.authTTL
	if tokenTTL <= 0 {
		tokenTTL = 30 * time.Minute
	}
	renewTTL := h.renewTTL
	if renewTTL <= 0 {
		renewTTL = 90 * 24 * time.Hour
	}
	authToken, _ := auth.GenerateToken(userID, tokenTTL)
	renewToken, _ := auth.GenerateToken(userID, renewTTL)
	writeJSON(w, http.StatusOK, map[string]any{"auth_token": authToken, "renew_token": renewToken})
}

func (h *APIHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		AuthToken string `json:"auth_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.AuthToken == "" {
		writeError(w, http.StatusBadRequest, "auth_token required")
		return
	}

	auth.RevokeToken(req.AuthToken)
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token, err := h.identitySvc.GenerateResetToken(req.Email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if h.mailer != nil {
		_ = h.mailer.SendReset(req.Email, token)
	}
	resp := map[string]any{"reset_token": token}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *APIHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.identitySvc.ResetPassword(req.Token, req.Password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	token := chi.URLParam(r, "token")
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.identitySvc.ConfirmEmail(token); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) VerifyRecaptcha(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if h.recaptchaSecret == "" {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", map[string][]string{
		"secret":   {h.recaptchaSecret},
		"response": {req.Token},
	})
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Success bool     `json:"success"`
		Errors  []string `json:"error-codes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !result.Success {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "errors": result.Errors})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}

func (h *APIHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rooms, err := h.roomSvc.List()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rooms)
}

func (h *APIHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Name    string `json:"name"`
		OwnerID string `json:"owner_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	room, err := h.roomSvc.Create(req.Name, req.OwnerID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, room)
}

func (h *APIHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.gameSvc == nil {
		writeError(w, http.StatusInternalServerError, "game service unavailable")
		return
	}
	var req struct {
		Slug    string `json:"slug"`
		OwnerID string `json:"owner_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Slug == "" || req.OwnerID == "" {
		writeError(w, http.StatusBadRequest, "slug and owner_id are required")
		return
	}
	gameUI, err := h.gameSvc.CreateGame(r.Context(), req.Slug, req.OwnerID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, gameUI)
}

func (h *APIHandler) ListPlugins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	plugins, err := h.pluginSvc.List()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(plugins)
}

func (h *APIHandler) ListVisiblePlugins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	plugins, err := h.pluginSvc.ListVisible()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(plugins)
}

func (h *APIHandler) CreatePlugin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Name    string `json:"name"`
		Visible bool   `json:"visible"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	plug, err := h.pluginSvc.Create(req.Name, req.Visible)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, plug)
}

func (h *APIHandler) GetPlugin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	pluginID := chi.URLParam(r, "pluginID")
	if pluginID == "" {
		writeError(w, http.StatusBadRequest, "plugin_id required")
		return
	}
	plug, err := h.pluginSvc.FindByID(pluginID)
	if err != nil {
		writeError(w, http.StatusNotFound, "plugin not found")
		return
	}
	writeJSON(w, http.StatusOK, plug)
}

func (h *APIHandler) ListCustomCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	pluginID := chi.URLParam(r, "pluginID")
	if pluginID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cards, err := h.pluginSvc.ListCustomCards(pluginID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(cards)
}

func (h *APIHandler) CreateCustomCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		// fallback: parse Bearer token directly when Auth middleware wasn't applied
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token := strings.TrimSpace(authHeader[len("bearer "):])
			if uid, err := auth.ParseToken(token); err == nil && uid != "" {
				userID = uid
			}
		}
	}
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	pluginID := chi.URLParam(r, "pluginID")
	if pluginID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	isAdmin, err := h.identitySvc.IsAdmin(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isAdmin {
		allowed, err := h.pluginSvc.HasPluginAccess(userID, pluginID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !allowed {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
	}
	var req struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	card, err := h.pluginSvc.CreateCustomCard(pluginID, req.Name, req.Data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(card)
}

func roomsRouter(h *APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListRooms(w, r)
		case http.MethodPost:
			h.CreateRoom(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func pluginsRouter(h *APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/be/api/plugins/visible" {
			h.ListVisiblePlugins(w, r)
			return
		}
		if r.Method == http.MethodGet {
			h.ListPlugins(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) CreateLfg(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		PluginID string `json:"plugin_id"`
		UserID   string `json:"user_id"`
		Text     string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	post, err := h.lfgSvc.Create(req.PluginID, req.UserID, req.Text)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(post)
}

func (h *APIHandler) ListLfg(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	pluginID := r.URL.Query().Get("plugin_id")
	posts, err := h.lfgSvc.List(pluginID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(posts)
}

func (h *APIHandler) DeleteLfg(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.lfgSvc.Delete(id); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) CreateAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Message string `json:"message"`
		Level   string `json:"level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	al, err := h.alertSvc.Create(req.Message, req.Level)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(al)
}

func (h *APIHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	alerts, err := h.alertSvc.List()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(alerts)
}

func (h *APIHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user, err := h.identitySvc.GetUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

func (h *APIHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user, err := h.identitySvc.GetUserByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Password != "" {
		if err := h.identitySvc.SetPassword(userID, req.Password); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	if req.Email != "" {
		user.Email = req.Email
		if err := h.identitySvc.UpdateUser(user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

func (h *APIHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := h.identitySvc.DeleteUser(userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.identitySvc.ListUsers()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}

func (h *APIHandler) PluginPermission(w http.ResponseWriter, r *http.Request) {
	pluginID := chi.URLParam(r, "pluginID")
	userID := chi.URLParam(r, "userID")
	if pluginID == "" || userID == "" {
		writeError(w, http.StatusBadRequest, "pluginID and userID are required")
		return
	}

	requesterID := middleware.GetUserIDFromContext(r.Context())
	if requesterID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	isAdmin, err := h.identitySvc.IsAdmin(requesterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if r.Method == http.MethodGet {
		perm, err := h.pluginSvc.GetUserPluginPermission(pluginID, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, perm)
		return
	}

	if r.Method == http.MethodPost || r.Method == http.MethodDelete {
		if !isAdmin {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	if r.Method == http.MethodPost {
		var req struct {
			Allowed bool `json:"allowed"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json")
			return
		}
		perm, err := h.pluginSvc.SetUserPluginPermission(pluginID, userID, req.Allowed)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, perm)
		return
	}

	if r.Method == http.MethodDelete {
		if err := h.pluginSvc.DeleteUserPluginPermission(pluginID, userID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (h *APIHandler) ListVisiblePluginsByUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "userID is required")
		return
	}
	plugins, err := h.pluginSvc.ListVisible()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user_id": userID, "plugins": plugins})
}

func (h *APIHandler) ListVisiblePluginByIDAndUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	pluginID := chi.URLParam(r, "pluginID")
	userID := chi.URLParam(r, "userID")
	if pluginID == "" || userID == "" {
		writeError(w, http.StatusBadRequest, "pluginID and userID are required")
		return
	}
	pluginObj, err := h.pluginSvc.FindByID(pluginID)
	if err != nil {
		writeError(w, http.StatusNotFound, "plugin not found")
		return
	}
	if !pluginObj.Visible {
		writeError(w, http.StatusForbidden, "plugin not visible")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user_id": userID, "plugin": pluginObj})
}

func (h *APIHandler) PluginRepoUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		RepoURL string `json:"repo_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.RepoURL == "" {
		writeError(w, http.StatusBadRequest, "repo_url required")
		return
	}
	if h.pluginSvc == nil {
		writeError(w, http.StatusInternalServerError, "plugin service unavailable")
		return
	}
	plugins, err := h.pluginSvc.SyncRepository(req.RepoURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "synced", "count": len(plugins), "plugins": plugins})
}

func (h *APIHandler) GetSetting(w http.ResponseWriter, r *http.Request) {
	if h.settingsSvc == nil {
		writeError(w, http.StatusInternalServerError, "settings service unavailable")
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	requesterID := middleware.GetUserIDFromContext(r.Context())
	if requesterID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	userID := chi.URLParam(r, "userID")
	pluginID := chi.URLParam(r, "pluginID")
	if userID == "" || pluginID == "" {
		writeError(w, http.StatusBadRequest, "userID and pluginID required")
		return
	}
	setting, err := h.settingsSvc.Get(userID, pluginID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, setting)
}

func (h *APIHandler) UpsertSetting(w http.ResponseWriter, r *http.Request) {
	if h.settingsSvc == nil {
		writeError(w, http.StatusInternalServerError, "settings service unavailable")
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	requesterID := middleware.GetUserIDFromContext(r.Context())
	if requesterID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req settings.Setting
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	setting, err := h.settingsSvc.Upsert(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, setting)
}

func (h *APIHandler) DeleteSetting(w http.ResponseWriter, r *http.Request) {
	if h.settingsSvc == nil {
		writeError(w, http.StatusInternalServerError, "settings service unavailable")
		return
	}
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	requesterID := middleware.GetUserIDFromContext(r.Context())
	if requesterID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	userID := chi.URLParam(r, "userID")
	pluginID := chi.URLParam(r, "pluginID")
	if userID == "" || pluginID == "" {
		writeError(w, http.StatusBadRequest, "userID and pluginID required")
		return
	}
	if err := h.settingsSvc.Delete(userID, pluginID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) AdminContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "message": "admin contact endpoint"})
		return
	}

	var req struct {
		Email   string `json:"email"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Email == "" || req.Message == "" {
		writeError(w, http.StatusBadRequest, "email and message required")
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{"status": "queued"})
}

func (h *APIHandler) AdminUpdateUserPatreon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		UserID string `json:"user_id"`
		Tier   string `json:"tier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.UserID == "" || req.Tier == "" {
		writeError(w, http.StatusBadRequest, "user_id and tier required")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user_id": req.UserID, "tier": req.Tier, "updated": true})
}

func (h *APIHandler) CreateDeck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req struct {
		ID       string          `json:"id"`
		Name     string          `json:"name"`
		PluginID string          `json:"plugin_id"`
		Data     json.RawMessage `json:"data"`
		Public   bool            `json:"public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	deckObj, err := deckDomain.NewDeck(req.ID, userID, req.PluginID, req.Name, req.Data, req.Public)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.deckSvc.Create(deckObj); err != nil {
		if err.Error() == "deck already exists" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(deckObj)
}

func (h *APIHandler) ListDecks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	decks, err := h.deckSvc.ListByUser(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(decks)
}

func (h *APIHandler) GetDeck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	currentUser := middleware.GetUserIDFromContext(r.Context())
	if currentUser == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	deckObj, err := h.deckSvc.GetByID(deckID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if deckObj.UserID != currentUser && !deckObj.Public {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(deckObj)
}

func (h *APIHandler) UpdateDeck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	currentUser := middleware.GetUserIDFromContext(r.Context())
	if currentUser == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	existing, err := h.deckSvc.GetByID(deckID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if existing.UserID != currentUser {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var req struct {
		Name   string          `json:"name"`
		Data   json.RawMessage `json:"data"`
		Public bool            `json:"public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	existing.Name = req.Name
	existing.Data = req.Data
	existing.Public = req.Public
	existing.Updated = time.Now().UTC()
	if err := h.deckSvc.Update(existing); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(existing)
}

func (h *APIHandler) DeleteDeck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	currentUser := middleware.GetUserIDFromContext(r.Context())
	if currentUser == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	deckID := chi.URLParam(r, "deckID")
	if deckID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	existing, err := h.deckSvc.GetByID(deckID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if existing.UserID != currentUser {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if err := h.deckSvc.Delete(deckID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) ListPublicDecks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	pluginID := chi.URLParam(r, "pluginID")
	if pluginID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	decks, err := h.deckSvc.ListPublicByPlugin(pluginID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(decks)
}

func (h *APIHandler) CreateReplay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var req struct {
		ID       string          `json:"id"`
		RoomSlug string          `json:"room_slug"`
		Meta     json.RawMessage `json:"meta"`
		Data     json.RawMessage `json:"data"`
		Public   bool            `json:"public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	replayObj, err := replayDomain.NewReplay(req.ID, req.RoomSlug, userID, req.Meta, req.Data, req.Public)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.replaySvc.Save(replayObj); err != nil {
		if err.Error() == "replay already exists" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(replayObj)
}

func (h *APIHandler) ListReplays(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	roomSlug := r.URL.Query().Get("room_slug")
	var result []*replayDomain.Replay
	var err error
	if roomSlug != "" {
		result, err = h.replaySvc.ListByRoom(roomSlug)
	} else {
		result, err = h.replaySvc.ListByOwner(userID)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (h *APIHandler) GetReplay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	replayID := chi.URLParam(r, "replayID")
	if replayID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	replayObj, err := h.replaySvc.GetByID(replayID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if replayObj.OwnerID != userID && !replayObj.IsPublic {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(replayObj)
}

func (h *APIHandler) DeleteReplay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	replayID := chi.URLParam(r, "replayID")
	if replayID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	replayObj, err := h.replaySvc.GetByID(replayID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if replayObj.OwnerID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if err := h.replaySvc.Delete(replayID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func NewRouter(h *APIHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)

	r.Get("/be/health", h.Health)
	r.Post("/be/api/v1/registration", h.Register)
	r.Post("/be/api/v1/session", h.Login)
	r.Delete("/be/api/v1/session", h.Logout)
	r.Post("/be/api/v1/session/renew", h.RenewSession)
	r.Get("/be/api/v1/confirm-email/{token}", h.ConfirmEmail)
	r.Get("/be/api/v1/confirm-email", h.ConfirmEmail) // legacy query param support
	r.Post("/be/api/v1/reset-password", h.RequestPasswordReset)
	r.Post("/be/api/v1/reset-password/update", h.ResetPassword)
	r.Post("/be/api/v1/recaptcha/verify", h.VerifyRecaptcha)

	r.Route("/be/api", func(r chi.Router) {
		// Old legacy routes remain for compatibility
		r.Handle("/rooms", roomsRouter(h))
		r.Handle("/plugins", pluginsRouter(h))
		r.Get("/plugins/visible", h.ListVisiblePlugins)
		r.Get("/plugins/visible/{userID}", h.ListVisiblePluginsByUser)
		r.Get("/plugins/visible/{pluginID}/{userID}", h.ListVisiblePluginByIDAndUser)
		r.Post("/plugin-repo-update", h.PluginRepoUpdate)
		r.Get("/admin_contact", h.AdminContact)
		r.Post("/admin_contact", h.AdminContact)
		r.Post("/plugins", h.CreatePlugin)
		r.Route("/plugins/{pluginID}", func(r chi.Router) {
			r.Get("/", h.GetPlugin)
			r.Get("/cards", h.ListCustomCards)
			r.With(middleware.Auth).Post("/cards", h.CreateCustomCard)
		})
		r.With(middleware.Auth).Route("/lfg", func(r chi.Router) {
			r.Get("/", h.ListLfg)
			r.Post("/", h.CreateLfg)
			r.Delete("/", h.DeleteLfg)
		})
		r.With(middleware.Auth).Route("/alerts", func(r chi.Router) {
			r.Get("/", h.ListAlerts)
			r.Post("/", h.CreateAlert)
		})
	})

	r.Route("/be/api/v1", func(r chi.Router) {
		r.Handle("/rooms", roomsRouter(h))
		r.Handle("/plugins", pluginsRouter(h))
		r.Get("/plugins/visible", h.ListVisiblePlugins)
		r.Get("/plugins/visible/{userID}", h.ListVisiblePluginsByUser)
		r.Get("/plugins/visible/{pluginID}/{userID}", h.ListVisiblePluginByIDAndUser)
		r.Post("/plugin-repo-update", h.PluginRepoUpdate)
		r.Post("/plugins", h.CreatePlugin)
		r.Route("/plugins/{pluginID}", func(r chi.Router) {
			r.Get("/", h.GetPlugin)
			r.Get("/cards", h.ListCustomCards)
			r.With(middleware.Auth).Post("/cards", h.CreateCustomCard)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth)
			r.Get("/profile", h.GetProfile)
			r.Post("/profile", h.UpdateProfile)
			r.Delete("/profile", h.DeleteProfile)
			r.Get("/users/all", h.ListUsers)
		})

		r.Get("/admin_contact", h.AdminContact)
		r.Post("/admin_contact", h.AdminContact)
		r.Post("/admin/update_user_patreon", h.AdminUpdateUserPatreon)

		r.With(middleware.Auth).Route("/users/plugin_permission/{pluginID}", func(r chi.Router) {
			r.Get("/{userID}", h.PluginPermission)
			r.Post("/{userID}", h.PluginPermission)
			r.Delete("/{userID}", h.PluginPermission)
		})

		r.Post("/games", h.CreateGame)

		r.With(middleware.Auth).Route("/settings", func(r chi.Router) {
			r.Post("/", h.UpsertSetting)
			r.Get("/{userID}/{pluginID}", h.GetSetting)
			r.Delete("/{userID}/{pluginID}", h.DeleteSetting)
		})

		r.Route("/decks", func(r chi.Router) {
			r.Get("/", h.ListDecks)
			r.Post("/", h.CreateDeck)
			r.Route("/{deckID}", func(r chi.Router) {
				r.Get("/", h.GetDeck)
				r.Put("/", h.UpdateDeck)
				r.Delete("/", h.DeleteDeck)
			})
		})
		r.Get("/public_decks/{pluginID}", h.ListPublicDecks)

		r.Route("/replays", func(r chi.Router) {
			r.Get("/", h.ListReplays)
			r.Post("/", h.CreateReplay)
			r.Route("/{replayID}", func(r chi.Router) {
				r.Get("/", h.GetReplay)
				r.Delete("/", h.DeleteReplay)
			})
		})

		r.Route("/lfg", func(r chi.Router) {
			r.Get("/", h.ListLfg)
			r.Post("/", h.CreateLfg)
			r.Delete("/", h.DeleteLfg)
		})

		r.Route("/alerts", func(r chi.Router) {
			r.Get("/", h.ListAlerts)
			r.Post("/", h.CreateAlert)
		})
	})

	return r
}
