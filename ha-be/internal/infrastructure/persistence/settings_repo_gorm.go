package persistence

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ldmonster/dragncards/ha-be/internal/domain/settings"
)

type GormSettingsRepository struct {
	db *gorm.DB
}

func NewGormSettingsRepository(db *gorm.DB) *GormSettingsRepository {
	return &GormSettingsRepository{db: db}
}

func (r *GormSettingsRepository) CreateOrUpdate(setting *settings.Setting) error {
	if setting == nil || setting.UserID == "" || setting.PluginID == "" {
		return errors.New("invalid setting")
	}
	setting.ID = setting.UserID + ":" + setting.PluginID
	return r.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "plugin_id"}}, DoUpdates: clause.AssignmentColumns([]string{"card_alt", "card_back_alt", "background_alt"})}).Save(setting).Error
}

func (r *GormSettingsRepository) Get(userID, pluginID string) (*settings.Setting, error) {
	if userID == "" || pluginID == "" {
		return nil, errors.New("user_id and plugin_id required")
	}
	var setting settings.Setting
	if err := r.db.First(&setting, "user_id = ? AND plugin_id = ?", userID, pluginID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("setting not found")
		}
		return nil, err
	}
	return &setting, nil
}

func (r *GormSettingsRepository) ListByUser(userID string) ([]*settings.Setting, error) {
	if userID == "" {
		return nil, errors.New("user_id required")
	}
	var settingsList []*settings.Setting
	if err := r.db.Where("user_id = ?", userID).Find(&settingsList).Error; err != nil {
		return nil, err
	}
	return settingsList, nil
}

func (r *GormSettingsRepository) Delete(userID, pluginID string) error {
	if userID == "" || pluginID == "" {
		return errors.New("user_id and plugin_id required")
	}
	return r.db.Where("user_id = ? AND plugin_id = ?", userID, pluginID).Delete(&settings.Setting{}).Error
}
