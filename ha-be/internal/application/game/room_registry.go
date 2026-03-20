package game

import (
	"sync"

	domaingame "github.com/ldmonster/dragncards/ha-be/internal/domain/game"
)

// GameRoom is a per-room goroutine holder with an action inbox channel.
type GameRoom struct {
	Slug  string
	State *domaingame.GameUI
	In    chan []byte
	done  chan struct{} // closed by Close() to signal the run goroutine to stop
}

// Close signals the room goroutine to stop and closes the inbox channel.
// Safe to call multiple times.
func (r *GameRoom) Close() {
	select {
	case <-r.done:
		// already closed
	default:
		close(r.done)
	}
}

// Done returns a channel that is closed when the room has been shut down.
func (r *GameRoom) Done() <-chan struct{} {
	return r.done
}

// RoomRegistry is a thread-safe in-memory map of slug → *GameRoom.
type RoomRegistry struct {
	mu    sync.RWMutex
	rooms map[string]*GameRoom
}

func NewRoomRegistry() *RoomRegistry {
	return &RoomRegistry{rooms: map[string]*GameRoom{}}
}

func (r *RoomRegistry) Register(room *GameRoom) {
	if room == nil || room.Slug == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rooms[room.Slug] = room
}

func (r *RoomRegistry) Unregister(slug string) {
	if slug == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if room, ok := r.rooms[slug]; ok {
		room.Close()
		delete(r.rooms, slug)
	}
}

func (r *RoomRegistry) Get(slug string) *GameRoom {
	if slug == "" {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.rooms[slug]
}

// CloseAll shuts down all registered rooms (called on server shutdown).
func (r *RoomRegistry) CloseAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, room := range r.rooms {
		room.Close()
	}
}
