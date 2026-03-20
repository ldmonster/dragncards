package game

// Stack is an ordered pile of Card IDs that lives inside a Group.
// Cards at index 0 are at the bottom; the last index is the top.
type Stack struct {
	ID      string   `json:"id"`
	GroupID string   `json:"group_id"`
	CardIDs []string `json:"card_ids"`
	Order   int      `json:"order"` // display order within its Group
}

// NewStack creates an empty Stack belonging to the given group.
func NewStack(id, groupID string) *Stack {
	return &Stack{
		ID:      id,
		GroupID: groupID,
		CardIDs: []string{},
	}
}

// AddCard appends a card ID to the top of the stack.
func (s *Stack) AddCard(cardID string) {
	if cardID == "" {
		return
	}
	s.CardIDs = append(s.CardIDs, cardID)
}

// RemoveCard removes a card ID from the stack (first occurrence).
func (s *Stack) RemoveCard(cardID string) {
	for i, id := range s.CardIDs {
		if id == cardID {
			s.CardIDs = append(s.CardIDs[:i], s.CardIDs[i+1:]...)
			return
		}
	}
}

// TopCardID returns the ID of the topmost card or "" if empty.
func (s *Stack) TopCardID() string {
	if len(s.CardIDs) == 0 {
		return ""
	}
	return s.CardIDs[len(s.CardIDs)-1]
}
