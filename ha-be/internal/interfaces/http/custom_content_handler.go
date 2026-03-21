package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ldmonster/dragncards/ha-be/internal/platform/middleware"
)

func (h *APIHandler) ListCustomContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	pluginID := r.URL.Query().Get("plugin_id")
	if pluginID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cards, err := h.customContentSvc.ListByPlugin(pluginID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, cards)
}

func (h *APIHandler) CreateCustomContent(w http.ResponseWriter, r *http.Request) {
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
		PluginID string `json:"plugin_id"`
		Name     string `json:"name"`
		Data     string `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.PluginID == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "plugin_id and name required")
		return
	}
	content, err := h.customContentSvc.Create(req.PluginID, userID, req.Name, req.Data)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, content)
}

func (h *APIHandler) DeleteCustomContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	cardID := chi.URLParam(r, "cardID")
	if cardID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	content, err := h.customContentSvc.GetByID(cardID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if content.OwnerID != userID {
		isAdmin, err := h.identitySvc.IsAdmin(userID)
		if err != nil || !isAdmin {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}
	if err := h.customContentSvc.Delete(cardID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) ListMyCustomContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := chi.URLParam(r, "userID")
	pluginID := chi.URLParam(r, "pluginID")
	if userID == "" || pluginID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cards, err := h.customContentSvc.ListByOwner(userID, pluginID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, cards)
}

func (h *APIHandler) ListAllCustomContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	userID := chi.URLParam(r, "userID")
	pluginID := chi.URLParam(r, "pluginID")
	if userID == "" || pluginID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cards, err := h.customContentSvc.ListByPlugin(pluginID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user_id": userID, "plugin_id": pluginID, "cards": cards})
}
