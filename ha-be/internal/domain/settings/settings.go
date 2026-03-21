package settings

type Setting struct {
	ID            string `json:"id" gorm:"primaryKey;type:varchar(64)"`
	UserID        string `json:"user_id" gorm:"index;type:varchar(64);not null"`
	PluginID      string `json:"plugin_id" gorm:"index;type:varchar(64);not null"`
	CardAlt       string `json:"card_alt" gorm:"type:varchar(256);default:''"`
	CardBackAlt   string `json:"card_back_alt" gorm:"type:varchar(256);default:''"`
	BackgroundAlt string `json:"background_alt" gorm:"type:varchar(256);default:''"`
}

type SettingRepository interface {
	CreateOrUpdate(setting *Setting) error
	Get(userID, pluginID string) (*Setting, error)
	ListByUser(userID string) ([]*Setting, error)
	Delete(userID, pluginID string) error
}
