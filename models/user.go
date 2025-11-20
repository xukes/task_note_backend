package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Password    string    `gorm:"not null" json:"-"`
	TOTPSecret  string    `json:"-"`
	TOTPEnabled bool      `gorm:"default:false" json:"totp_enabled"`
	CreatedAt   time.Time `json:"created_at"`
}
