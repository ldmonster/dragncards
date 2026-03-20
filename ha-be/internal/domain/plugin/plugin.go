package plugin

type Plugin struct {
	ID      string `json:"id" gorm:"primaryKey;type:varchar(64)"`
	Name    string `json:"name" gorm:"type:varchar(256);not null"`
	Visible bool   `json:"visible" gorm:"not null;default:false"`
}
