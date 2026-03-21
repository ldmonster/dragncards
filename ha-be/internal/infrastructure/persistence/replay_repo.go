package persistence

import (
	"errors"
	"sync"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/replay"
)

var ErrReplayNotFound = errors.New("replay not found")

type InMemoryReplayRepository struct {
	mu      sync.RWMutex
	replays map[string]*replay.Replay
}

func NewInMemoryReplayRepository() *InMemoryReplayRepository {
	return &InMemoryReplayRepository{replays: make(map[string]*replay.Replay)}
}

func (r *InMemoryReplayRepository) Save(rp *replay.Replay) error {
	if rp == nil || rp.ID == "" {
		return errors.New("invalid replay")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.replays[rp.ID] = rp
	return nil
}

func (r *InMemoryReplayRepository) FindByID(id string) (*replay.Replay, error) {
	if id == "" {
		return nil, ErrReplayNotFound
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	rp, ok := r.replays[id]
	if !ok {
		return nil, ErrReplayNotFound
	}
	return rp, nil
}

func (r *InMemoryReplayRepository) FindByRoomSlug(roomSlug string) ([]*replay.Replay, error) {
	if roomSlug == "" {
		return nil, errors.New("room_slug required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	mapped := make([]*replay.Replay, 0)
	for _, rp := range r.replays {
		if rp.RoomSlug == roomSlug {
			mapped = append(mapped, rp)
		}
	}
	return mapped, nil
}

func (r *InMemoryReplayRepository) FindByOwner(ownerID string) ([]*replay.Replay, error) {
	if ownerID == "" {
		return nil, errors.New("owner_id required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	mapped := make([]*replay.Replay, 0)
	for _, rp := range r.replays {
		if rp.OwnerID == ownerID {
			mapped = append(mapped, rp)
		}
	}
	return mapped, nil
}

func (r *InMemoryReplayRepository) Delete(id string) error {
	if id == "" {
		return ErrReplayNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.replays[id]; !ok {
		return ErrReplayNotFound
	}
	delete(r.replays, id)
	return nil
}
