package persistence

import (
	"errors"
	"sync"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/room"
)

var ErrRoomNotFound = errors.New("room not found")

type InMemoryRoomRepository struct {
	mu      sync.RWMutex
	rooms   map[string]*room.Room
	actions map[string][][]byte
}

func NewInMemoryRoomRepository() *InMemoryRoomRepository {
	return &InMemoryRoomRepository{rooms: map[string]*room.Room{}, actions: map[string][][]byte{}}
}

func (r *InMemoryRoomRepository) Create(room *room.Room) error {
	if room == nil || room.Slug == "" {
		return errors.New("invalid room")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rooms[room.Slug]; exists {
		return errors.New("room already exists")
	}

	r.rooms[room.Slug] = room
	return nil
}

func (r *InMemoryRoomRepository) List() ([]*room.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rooms := make([]*room.Room, 0, len(r.rooms))
	for _, v := range r.rooms {
		rooms = append(rooms, v)
	}
	return rooms, nil
}

func (r *InMemoryRoomRepository) FindBySlug(slug string) (*room.Room, error) {
	if slug == "" {
		return nil, ErrRoomNotFound
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	room, ok := r.rooms[slug]
	if !ok {
		return nil, ErrRoomNotFound
	}
	return room, nil
}

func (r *InMemoryRoomRepository) AppendAction(slug string, payload []byte) error {
	if slug == "" || len(payload) == 0 {
		return errors.New("invalid action")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.rooms[slug]; !ok {
		return ErrRoomNotFound
	}

	r.actions[slug] = append(r.actions[slug], payload)
	return nil
}

func (r *InMemoryRoomRepository) ListActions(slug string) ([][]byte, error) {
	if slug == "" {
		return nil, ErrRoomNotFound
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.rooms[slug]; !ok {
		return nil, ErrRoomNotFound
	}

	actions := make([][]byte, len(r.actions[slug]))
	copy(actions, r.actions[slug])
	return actions, nil
}
