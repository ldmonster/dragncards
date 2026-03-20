package persistence

import (
	"errors"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/deck"
	"gorm.io/gorm"
)

type GormDeckRepository struct {
	db *gorm.DB
}

func NewGormDeckRepository(db *gorm.DB) *GormDeckRepository {
	return &GormDeckRepository{db: db}
}

func (r *GormDeckRepository) Save(d *deck.Deck) error {
	if d == nil || d.ID == "" {
		return errors.New("invalid deck")
	}
	return r.db.Save(d).Error
}

func (r *GormDeckRepository) FindByID(id string) (*deck.Deck, error) {
	if id == "" {
		return nil, ErrDeckNotFound
	}
	var d deck.Deck
	if err := r.db.First(&d, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeckNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *GormDeckRepository) FindByUser(userID string) ([]*deck.Deck, error) {
	if userID == "" {
		return nil, errors.New("user_id required")
	}
	var decks []*deck.Deck
	if err := r.db.Where("user_id = ?", userID).Find(&decks).Error; err != nil {
		return nil, err
	}
	return decks, nil
}

func (r *GormDeckRepository) FindPublicByPlugin(pluginID string) ([]*deck.Deck, error) {
	if pluginID == "" {
		return nil, errors.New("plugin_id required")
	}
	var decks []*deck.Deck
	if err := r.db.Where("plugin_id = ? AND public = ?", pluginID, true).Find(&decks).Error; err != nil {
		return nil, err
	}
	return decks, nil
}

func (r *GormDeckRepository) Delete(id string) error {
	if id == "" {
		return ErrDeckNotFound
	}
	result := r.db.Delete(&deck.Deck{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrDeckNotFound
	}
	return nil
}
