package identity

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	identityapp "github.com/ldmonster/dragncards/ha-be/internal/application/identity"
	"github.com/ldmonster/dragncards/ha-be/internal/domain/identity"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/persistence"
)

func newTestIdentityHandler(t *testing.T) *Handler {
	domainSvc := identity.NewService(persistence.NewInMemoryUserRepository())
	appSvc := identityapp.NewService(domainSvc)
	appSvc.SetTokenTTL(1*time.Hour, 1*time.Hour)
	return NewHandler(appSvc, nil, "", 30*time.Minute, 90*24*time.Hour)
}

func TestIdentityHandlerRegisterLoginConfirmReset(t *testing.T) {
	h := newTestIdentityHandler(t)

	// Register
	registerBody := map[string]string{"email": "test@id.example", "password": "p@ssw0rd"}
	b, _ := json.Marshal(registerBody)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/be/api/v1/registration", bytes.NewReader(b))
	h.Register(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rr.Code)
	}
	var regResp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("invalid register response JSON: %v", err)
	}
	if regResp["email"] != "test@id.example" {
		t.Fatalf("expected email in response, got %v", regResp["email"])
	}
	confirmToken, ok := regResp["confirmation_token"].(string)
	if !ok || confirmToken == "" {
		t.Fatal("expected confirmation token in register response")
	}

	// Confirm email via path param
	confirmRR := httptest.NewRecorder()
	confirmReq := httptest.NewRequest(http.MethodGet, "/be/api/v1/confirm-email/"+confirmToken, nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("token", confirmToken)
	confirmReq = confirmReq.WithContext(context.WithValue(confirmReq.Context(), chi.RouteCtxKey, ctx))
	h.ConfirmEmail(confirmRR, confirmReq)
	if confirmRR.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for confirm email, got %d", confirmRR.Code)
	}

	// Login
	loginRR := httptest.NewRecorder()
	loginBody := map[string]string{"email": "test@id.example", "password": "p@ssw0rd"}
	lb, _ := json.Marshal(loginBody)
	loginReq := httptest.NewRequest(http.MethodPost, "/be/api/v1/session", bytes.NewReader(lb))
	h.Login(loginRR, loginReq)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for login, got %d", loginRR.Code)
	}
	var loginResp map[string]any
	if err := json.Unmarshal(loginRR.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("invalid login response JSON: %v", err)
	}
	if loginResp["auth_token"] == "" {
		t.Fatalf("expected auth_token in login response")
	}

	// Request password reset
	resetReqRR := httptest.NewRecorder()
	resetReqBody := map[string]string{"email": "test@id.example"}
	rb, _ := json.Marshal(resetReqBody)
	resetReq := httptest.NewRequest(http.MethodPost, "/be/api/v1/reset-password", bytes.NewReader(rb))
	h.RequestPasswordReset(resetReqRR, resetReq)
	if resetReqRR.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for reset request, got %d", resetReqRR.Code)
	}
	var resetResp map[string]any
	if err := json.Unmarshal(resetReqRR.Body.Bytes(), &resetResp); err != nil {
		t.Fatalf("invalid reset request response JSON: %v", err)
	}
	resetToken, ok := resetResp["reset_token"].(string)
	if !ok || resetToken == "" {
		t.Fatal("expected reset_token in response")
	}

	// Perform password reset
	resetRR := httptest.NewRecorder()
	resetBody := map[string]string{"token": resetToken, "password": "n3wp@ss"}
	rb2, _ := json.Marshal(resetBody)
	resetReq2 := httptest.NewRequest(http.MethodPost, "/be/api/v1/reset-password/update", bytes.NewReader(rb2))
	h.ResetPassword(resetRR, resetReq2)
	if resetRR.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for reset password, got %d", resetRR.Code)
	}

	// Login with new password
	loginRR2 := httptest.NewRecorder()
	loginBody2 := map[string]string{"email": "test@id.example", "password": "n3wp@ss"}
	lb2, _ := json.Marshal(loginBody2)
	loginReq2 := httptest.NewRequest(http.MethodPost, "/be/api/v1/session", bytes.NewReader(lb2))
	h.Login(loginRR2, loginReq2)
	if loginRR2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for login after reset, got %d", loginRR2.Code)
	}
}
