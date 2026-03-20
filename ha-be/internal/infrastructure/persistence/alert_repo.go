package persistence

import (
	"errors"
	"sync"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/alert"
)

var ErrAlertNotFound = errors.New("alert not found")

type InMemoryAlertRepository struct {
	mu     sync.RWMutex
	alerts map[string]*alert.Alert
}

func NewInMemoryAlertRepository() *InMemoryAlertRepository {
	return &InMemoryAlertRepository{alerts: map[string]*alert.Alert{}}
}

func (r *InMemoryAlertRepository) Create(a *alert.Alert) error {
	if a == nil || a.ID == "" {
		return errors.New("invalid alert")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.alerts[a.ID]; ok {
		return errors.New("alert already exists")
	}

	r.alerts[a.ID] = a
	return nil
}

func (r *InMemoryAlertRepository) List() ([]*alert.Alert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*alert.Alert, 0, len(r.alerts))
	for _, a := range r.alerts {
		result = append(result, a)
	}
	return result, nil
}
