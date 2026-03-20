package deck

import (
	"errors"
	"time"
)

// Deck represents a user-owned card deck scoped to a plugin.
type Deck struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	PluginID string    `json:"plugin_id"`
	Name     string    `json:"name"`
	Data     []byte    `json:"data"` // opaque JSON from client
	Public   bool      `json:"public"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

func NewDeck(id, userID, pluginID, name string, data []byte, public bool) (*Deck, error) {
	if id == "" {
		return nil, errors.New("deck: id is required")
	}
	if userID == "" {
		return nil, errors.New("deck: user_id is required")
	}
	if pluginID == "" {
		return nil, errors.New("deck: plugin_id is required")
	}
	if name == "" {
		return nil, errors.New("deck: name is required")
	}
	now := time.Now().UTC()
	return &Deck{
		ID:       id,
		UserID:   userID,
		PluginID: pluginID,
		Name:     name,
		Data:     data,
		Public:   public,
		Created:  now,
		Updated:  now,
	}, nil
}
