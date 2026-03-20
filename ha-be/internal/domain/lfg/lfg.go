package lfg

type LfgPost struct {
	ID       string `json:"id" gorm:"primaryKey;type:varchar(64)"`
	PluginID string `json:"plugin_id" gorm:"type:varchar(64);not null"`
	UserID   string `json:"user_id" gorm:"type:varchar(64);not null"`
	Text     string `json:"text" gorm:"type:text;not null"`
}
