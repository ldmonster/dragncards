package customcontent
package custom_content

// CustomContent represents a user-created card-like artifact for a plugin.







}	Data     string `json:"data" gorm:"type:jsonb"`	Name     string `json:"name" gorm:"type:varchar(256);not null"`	OwnerID  string `json:"owner_id" gorm:"index;type:varchar(64);not null"`	PluginID string `json:"plugin_id" gorm:"index;type:varchar(64);not null"`	ID       string `json:"id" gorm:"primaryKey;type:varchar(64)"`type CustomContent struct {