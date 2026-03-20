package deck

// DeckRepository defines persistence methods for deck.
type DeckRepository interface {
	Save(deck *Deck) error
	FindByID(id string) (*Deck, error)
	FindByUser(userID string) ([]*Deck, error)
	FindPublicByPlugin(pluginID string) ([]*Deck, error)
	Delete(id string) error
}
