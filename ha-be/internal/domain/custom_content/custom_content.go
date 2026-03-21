package custom_content

// CustomContent represents a user-created card-like artifact for a plugin.
type CustomContent struct {
	ID       string `json:"id" gorm:"primaryKey;type:varchar(64)"`
	PluginID string `json:"plugin_id" gorm:"index;type:varchar(64);not null"`
	OwnerID  string `json:"owner_id" gorm:"index;type:varchar(64);not null"`
	Name     string `json:"name" gorm:"type:varchar(256);not null"`
	Data     string `json:"data" gorm:"type:jsonb"`
}
