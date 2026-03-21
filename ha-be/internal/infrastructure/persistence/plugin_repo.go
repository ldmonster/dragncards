package persistence

import (
	"errors"
	"sync"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/plugin"
)

var ErrPluginNotFound = errors.New("plugin not found")

type InMemoryPluginRepository struct {
	mu          sync.RWMutex
	plugins     map[string]*plugin.Plugin
	customCards map[string]*plugin.CustomCard
	permissions map[string]*plugin.UserPluginPermission
}

func NewInMemoryPluginRepository() *InMemoryPluginRepository {
	return &InMemoryPluginRepository{
		plugins:     map[string]*plugin.Plugin{},
		customCards: map[string]*plugin.CustomCard{},
		permissions: map[string]*plugin.UserPluginPermission{},
	}
}

func (r *InMemoryPluginRepository) Create(p *plugin.Plugin) error {
	if p == nil || p.ID == "" {
		return errors.New("invalid plugin")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.plugins[p.ID]; ok {
		return errors.New("plugin already exists")
	}

	r.plugins[p.ID] = p
	return nil
}

func (r *InMemoryPluginRepository) List() ([]*plugin.Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*plugin.Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		result = append(result, p)
	}
	return result, nil
}

func (r *InMemoryPluginRepository) ListVisible() ([]*plugin.Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*plugin.Plugin, 0)
	for _, p := range r.plugins {
		if p.Visible {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *InMemoryPluginRepository) FindByID(id string) (*plugin.Plugin, error) {
	if id == "" {
		return nil, ErrPluginNotFound
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.plugins[id]
	if !ok {
		return nil, ErrPluginNotFound
	}
	return p, nil
}

func (r *InMemoryPluginRepository) CreateCustomCard(card *plugin.CustomCard) error {
	if card == nil || card.ID == "" || card.PluginID == "" {
		return errors.New("invalid custom card")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.customCards[card.ID]; exists {
		return errors.New("custom card already exists")
	}
	r.customCards[card.ID] = card
	return nil
}

func (r *InMemoryPluginRepository) ListCustomCards(pluginID string) ([]*plugin.CustomCard, error) {
	if pluginID == "" {
		return nil, errors.New("plugin_id required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*plugin.CustomCard, 0)
	for _, c := range r.customCards {
		if c.PluginID == pluginID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (r *InMemoryPluginRepository) FindCustomCardByID(id string) (*plugin.CustomCard, error) {
	if id == "" {
		return nil, errors.New("custom card not found")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.customCards[id]
	if !ok {
		return nil, errors.New("custom card not found")
	}
	return c, nil
}

func (r *InMemoryPluginRepository) CreatePermission(permission *plugin.UserPluginPermission) error {
	if permission == nil || permission.PluginID == "" || permission.UserID == "" {
		return errors.New("invalid permission")
	}

	key := permission.PluginID + ":" + permission.UserID

	r.mu.Lock()
	defer r.mu.Unlock()

	r.permissions[key] = permission
	return nil
}

func (r *InMemoryPluginRepository) GetPermission(pluginID, userID string) (*plugin.UserPluginPermission, error) {
	if pluginID == "" || userID == "" {
		return nil, errors.New("permission lookup requires plugin_id and user_id")
	}

	key := pluginID + ":" + userID

	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.permissions[key]
	if !ok {
		return &plugin.UserPluginPermission{PluginID: pluginID, UserID: userID, Allowed: false}, nil
	}
	return p, nil
}

func (r *InMemoryPluginRepository) DeletePermission(pluginID, userID string) error {
	if pluginID == "" || userID == "" {
		return errors.New("permission delete requires plugin_id and user_id")
	}

	key := pluginID + ":" + userID

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.permissions, key)
	return nil
}
