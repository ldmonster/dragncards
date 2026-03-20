package game

// PlayerInfo holds the lobby / seating metadata for one connected player.
type PlayerInfo struct {
	ID        string `json:"id"`
	Alias     string `json:"alias"`
	Seat      int    `json:"seat"`      // 1-based seat number; 0 = unseated
	Spectator bool   `json:"spectator"` // true if watching only
}

// NewPlayerInfo creates a PlayerInfo entry for a joining user.
func NewPlayerInfo(id, alias string, seat int) *PlayerInfo {
	return &PlayerInfo{
		ID:    id,
		Alias: alias,
		Seat:  seat,
	}
}
