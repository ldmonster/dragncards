package replay

// ReplayRepository defines persistence operations for Replay aggregates.
type ReplayRepository interface {
	Save(replay *Replay) error
	FindByID(id string) (*Replay, error)
	FindByRoomSlug(roomSlug string) ([]*Replay, error)
	FindByOwner(ownerID string) ([]*Replay, error)
	Delete(id string) error
}
