package identity

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	identityapp "github.com/ldmonster/dragncards/ha-be/internal/application/identity"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/identity"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/email"
	"github.com/ldmonster/dragncards/ha-be/internal/platform/auth"
)

type Handler struct {
	identitySvc     *identityapp.Service
	mailer          email.Mailer
	recaptchaSecret string
	authTTL         time.Duration
	renewTTL        time.Duration
}

func NewHandler(identitySvc *identityapp.Service, mailer email.Mailer, recaptchaSecret string, authTTL, renewTTL time.Duration) *Handler {
	return &Handler{
		identitySvc:     identitySvc,
		mailer:          mailer,
		recaptchaSecret: recaptchaSecret,
		authTTL:         authTTL,
		renewTTL:        renewTTL,
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) RenewSession(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) ConfirmEmail(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) VerifyRecaptcha(w http.ResponseWriter, r *http.Request) {
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "errors": result.Errors})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
}
