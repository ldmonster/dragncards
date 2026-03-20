package game

// Card represents a single physical card in the game state.
// It belongs to a Stack (which in turn belongs to a Group).
type Card struct {
	ID        string         `json:"id"`
	StackID   string         `json:"stack_id"`
	SideUp    string         `json:"side_up"`   // "A" or "B"
	Rotation  int            `json:"rotation"`  // degrees: 0, 90, 180, 270
	Tokens    map[string]int `json:"tokens"`    // named token counters
	Exhausted bool           `json:"exhausted"` // tapped / rotated state
	Hidden    bool           `json:"hidden"`    // face-down / hidden from others
}

// NewCard creates a Card with its ID and owning stack slug.
func NewCard(id, stackID string) *Card {
	return &Card{
		ID:      id,
		StackID: stackID,
		SideUp:  "A",
		Tokens:  map[string]int{},
	}
}

// SetToken updates a named token counter on the card.
func (c *Card) SetToken(name string, value int) {
	if c.Tokens == nil {
		c.Tokens = map[string]int{}
	}
	c.Tokens[name] = value
}

// GetToken returns the current value of a named token (0 if absent).
func (c *Card) GetToken(name string) int {
	if c.Tokens == nil {
		return 0
	}
	return c.Tokens[name]
}
