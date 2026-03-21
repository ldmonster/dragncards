package persistence

import (
	"errors"
	"sync"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/deck"
)

var ErrDeckNotFound = errors.New("deck not found")

type InMemoryDeckRepository struct {
	mu    sync.RWMutex
	decks map[string]*deck.Deck
}

func NewInMemoryDeckRepository() *InMemoryDeckRepository {
	return &InMemoryDeckRepository{decks: make(map[string]*deck.Deck)}
}

func (r *InMemoryDeckRepository) Save(d *deck.Deck) error {
	if d == nil || d.ID == "" {
		return errors.New("invalid deck")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.decks[d.ID] = d
	return nil
}

func (r *InMemoryDeckRepository) FindByID(id string) (*deck.Deck, error) {
	if id == "" {
		return nil, ErrDeckNotFound
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	d, ok := r.decks[id]
	if !ok {
		return nil, ErrDeckNotFound
	}
	return d, nil
}

func (r *InMemoryDeckRepository) FindByUser(userID string) ([]*deck.Deck, error) {
	if userID == "" {
		return nil, errors.New("user_id required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*deck.Deck, 0)
	for _, d := range r.decks {
		if d.UserID == userID {
			result = append(result, d)
		}
	}
	return result, nil
}

func (r *InMemoryDeckRepository) FindPublicByPlugin(pluginID string) ([]*deck.Deck, error) {
	if pluginID == "" {
		return nil, errors.New("plugin_id required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*deck.Deck, 0)
	for _, d := range r.decks {
		if d.PluginID == pluginID && d.Public {
			result = append(result, d)
		}
	}
	return result, nil
}

func (r *InMemoryDeckRepository) Delete(id string) error {
	if id == "" {
		return ErrDeckNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.decks[id]; !ok {
		return ErrDeckNotFound
	}
	delete(r.decks, id)
	return nil
}
