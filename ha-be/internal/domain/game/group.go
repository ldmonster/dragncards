package game

// Group is a named zone that holds an ordered collection of Stacks.
// Each group has an optional owner (player ID); empty means shared/table zone.
type Group struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	OwnerID  string   `json:"owner_id"`  // empty = shared zone
	StackIDs []string `json:"stack_ids"` // ordered list of stack IDs
}

// NewGroup creates an empty Group with a label and optional owner.
func NewGroup(id, label, ownerID string) *Group {
	return &Group{
		ID:       id,
		Label:    label,
		OwnerID:  ownerID,
		StackIDs: []string{},
	}
}

// AddStack appends a stack ID to the group's ordered list.
func (g *Group) AddStack(stackID string) {
	if stackID == "" {
		return
	}
	g.StackIDs = append(g.StackIDs, stackID)
}

// RemoveStack removes a stack ID from the group (first occurrence).
func (g *Group) RemoveStack(stackID string) {
	for i, id := range g.StackIDs {
		if id == stackID {
			g.StackIDs = append(g.StackIDs[:i], g.StackIDs[i+1:]...)
			return
		}
	}
}
