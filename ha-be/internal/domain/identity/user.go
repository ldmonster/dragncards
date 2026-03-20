package identity

import "time"

type User struct {
	ID                    string    `json:"id" gorm:"primaryKey;type:varchar(64)"`
	Email                 string    `json:"email" gorm:"uniqueIndex;type:varchar(256);not null"`
	PasswordHash          string    `json:"-" gorm:"type:text;not null"`
	Confirmed             bool      `json:"confirmed" gorm:"not null;default:false"`
	ConfirmToken          string    `json:"-" gorm:"type:varchar(256)"`
	ConfirmTokenExpiresAt time.Time `json:"-" gorm:"index"`
	ResetToken            string    `json:"-" gorm:"type:varchar(256)"`
	ResetTokenExpiresAt   time.Time `json:"-" gorm:"index"`
}
