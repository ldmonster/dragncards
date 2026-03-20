package persistence

import (
	"errors"
	"sync"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/lfg"
)

var ErrLfgNotFound = errors.New("lfg post not found")

type InMemoryLfgRepository struct {
	mu    sync.RWMutex
	posts map[string]*lfg.LfgPost
}

func NewInMemoryLfgRepository() *InMemoryLfgRepository {
	return &InMemoryLfgRepository{posts: map[string]*lfg.LfgPost{}}
}

func (r *InMemoryLfgRepository) Create(post *lfg.LfgPost) error {
	if post == nil || post.ID == "" {
		return errors.New("invalid lfg post")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.posts[post.ID]; exists {
		return errors.New("lfg post already exists")
	}

	r.posts[post.ID] = post
	return nil
}

func (r *InMemoryLfgRepository) List(pluginID string) ([]*lfg.LfgPost, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*lfg.LfgPost, 0)
	for _, p := range r.posts {
		if pluginID == "" || p.PluginID == pluginID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (r *InMemoryLfgRepository) Delete(id string) error {
	if id == "" {
		return ErrLfgNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.posts[id]; !ok {
		return ErrLfgNotFound
	}
	delete(r.posts, id)
	return nil
}
