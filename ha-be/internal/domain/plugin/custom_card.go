package plugin

type CustomCard struct {
	ID       string `json:"id" gorm:"primaryKey;type:varchar(64)"`
	PluginID string `json:"plugin_id" gorm:"index;type:varchar(64);not null"`
	Name     string `json:"name" gorm:"type:varchar(256);not null"`
	Data     string `json:"data" gorm:"type:jsonb"`
}
