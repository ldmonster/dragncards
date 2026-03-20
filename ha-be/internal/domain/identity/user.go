package identity

type User struct {
	ID           string `json:"id" gorm:"primaryKey;type:varchar(64)"`
	Email        string `json:"email" gorm:"uniqueIndex;type:varchar(256);not null"`
	PasswordHash string `json:"-" gorm:"type:text;not null"`
	Confirmed    bool   `json:"confirmed" gorm:"not null;default:false"`
}
