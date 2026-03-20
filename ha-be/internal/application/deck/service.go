package deck

import (
	"errors"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/deck"
)

// DeckService performs application use cases for decks.
type DeckService struct {
	repo deck.DeckRepository
}

func NewDeckService(repo deck.DeckRepository) *DeckService {
	return &DeckService{repo: repo}
}

func (s *DeckService) Create(deckObj *deck.Deck) error {
	_, err := s.repo.FindByID(deckObj.ID)
	if err == nil {
		return errors.New("deck already exists")
	}
	if err := s.repo.Save(deckObj); err != nil {
		return err
	}
	return nil
}

func (s *DeckService) Update(deckObj *deck.Deck) error {
	if _, err := s.repo.FindByID(deckObj.ID); err != nil {
		return err
	}
	return s.repo.Save(deckObj)
}

func (s *DeckService) GetByID(id string) (*deck.Deck, error) {
	return s.repo.FindByID(id)
}

func (s *DeckService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *DeckService) ListByUser(userID string) ([]*deck.Deck, error) {
	return s.repo.FindByUser(userID)
}

func (s *DeckService) ListPublicByPlugin(pluginID string) ([]*deck.Deck, error) {
	return s.repo.FindPublicByPlugin(pluginID)
}
