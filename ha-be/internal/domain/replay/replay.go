package replay

import (
	"errors"
	"time"
)

// Replay represents a saved game replay
type Replay struct {
	ID       string    `json:"id"`
	RoomSlug string    `json:"room_slug"`
	OwnerID  string    `json:"owner_id"`
	Meta     []byte    `json:"meta"` // additional metadata JSON
	Data     []byte    `json:"data"` // replay events payload
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
	IsPublic bool      `json:"public"`
}

func NewReplay(id, roomSlug, ownerID string, meta, data []byte, isPublic bool) (*Replay, error) {
	if id == "" {
		return nil, errors.New("replay: id is required")
	}
	if roomSlug == "" {
		return nil, errors.New("replay: room_slug is required")
	}
	if ownerID == "" {
		return nil, errors.New("replay: owner_id is required")
	}

	now := time.Now().UTC()
	return &Replay{
		ID:       id,
		RoomSlug: roomSlug,
		OwnerID:  ownerID,
		Meta:     meta,
		Data:     data,
		Created:  now,
		Updated:  now,
		IsPublic: isPublic,
	}, nil
}
