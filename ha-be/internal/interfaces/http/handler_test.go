package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/identity"
	"github.com/ldmonster/dragncards/ha-be/internal/infrastructure/persistence"
)

func TestAPIHandlerHealth(t *testing.T) {
	h := NewAPIHandlerLegacy(nil, nil, nil, nil, nil, nil, nil, nil)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/be/health", nil)

	h.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unexpected body JSON: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", resp["status"])
	}
}

func TestAPIHandlerRegister_InvalidJSON(t *testing.T) {
	identityService := identity.NewService(persistence.NewInMemoryUserRepository())
	h := NewAPIHandlerLegacy(identityService, nil, nil, nil, nil, nil, nil, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/be/register", bytes.NewReader([]byte("not-json")))

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp["error"] != "invalid json" {
		t.Fatalf("expected invalid json error, got %q", resp["error"])
	}
}

func TestAPIHandlerRegister_Success(t *testing.T) {
	identityService := identity.NewService(persistence.NewInMemoryUserRepository())
	h := NewAPIHandlerLegacy(identityService, nil, nil, nil, nil, nil, nil, nil)

	body := map[string]string{"email": "test@example.com", "password": "p@ssw0rd"}
	bb, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/be/register", bytes.NewReader(bb))

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if resp["email"] != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %v", resp["email"])
	}
	if resp["id"] == "" || resp["id"] == nil {
		t.Fatalf("expected non-empty id")
	}
	if _, ok := resp["confirmation_token"]; !ok {
		t.Fatalf("expected confirmation_token in response")
	}
}

func TestAPIHandlerRegisterConflict(t *testing.T) {
	identityService := identity.NewService(persistence.NewInMemoryUserRepository())
	h := NewAPIHandlerLegacy(identityService, nil, nil, nil, nil, nil, nil, nil)

	// Register once successfully
	body := map[string]string{"email": "duplicate@example.com", "password": "p@ssw0rd"}
	bb, _ := json.Marshal(body)
	rr1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/be/register", bytes.NewReader(bb))
	h.Register(rr1, req1)
	if rr1.Code != http.StatusCreated {
		t.Fatalf("expected first registration 201 Created, got %d", rr1.Code)
	}

	// Register same email again and expect conflict
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/be/register", bytes.NewReader(bb))
	h.Register(rr2, req2)
	if rr2.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for duplicate registration, got %d", rr2.Code)
	}
}
