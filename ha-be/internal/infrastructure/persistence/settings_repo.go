package persistence

import (
	"errors"
	"sync"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/settings"
)

type InMemorySettingsRepository struct {
	mu       sync.RWMutex
	settings map[string]*settings.Setting
}

func NewInMemorySettingsRepository() *InMemorySettingsRepository {
	return &InMemorySettingsRepository{settings: make(map[string]*settings.Setting)}
}

func key(userID, pluginID string) string {
	return userID + ":" + pluginID
}

func (r *InMemorySettingsRepository) CreateOrUpdate(setting *settings.Setting) error {
	if setting == nil || setting.UserID == "" || setting.PluginID == "" {
		return errors.New("invalid setting")
	}
	k := key(setting.UserID, setting.PluginID)

	r.mu.Lock()
	defer r.mu.Unlock()
	setting.ID = k
	r.settings[k] = setting
	return nil
}

func (r *InMemorySettingsRepository) Get(userID, pluginID string) (*settings.Setting, error) {
	if userID == "" || pluginID == "" {
		return nil, errors.New("user_id and plugin_id required")
	}
	k := key(userID, pluginID)

	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.settings[k]
	if !ok {
		return nil, errors.New("setting not found")
	}
	return s, nil
}

func (r *InMemorySettingsRepository) ListByUser(userID string) ([]*settings.Setting, error) {
	if userID == "" {
		return nil, errors.New("user_id required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*settings.Setting, 0)
	for _, s := range r.settings {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *InMemorySettingsRepository) Delete(userID, pluginID string) error {
	if userID == "" || pluginID == "" {
		return errors.New("user_id and plugin_id required")
	}
	k := key(userID, pluginID)

	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.settings, k)
	return nil
}
