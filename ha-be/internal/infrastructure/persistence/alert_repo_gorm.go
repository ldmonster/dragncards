package persistence

import (
	"errors"

	"gorm.io/gorm"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/alert"
)

type GormAlertRepository struct {
	db *gorm.DB
}

func NewGormAlertRepository(db *gorm.DB) *GormAlertRepository {
	return &GormAlertRepository{db: db}
}

func (r *GormAlertRepository) Create(a *alert.Alert) error {
	if a == nil || a.ID == "" {
		return errors.New("invalid alert")
	}
	return r.db.Create(a).Error
}

func (r *GormAlertRepository) List() ([]*alert.Alert, error) {
	var alerts []*alert.Alert
	if err := r.db.Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}
