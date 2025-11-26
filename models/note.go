package models

import "time"

type Note struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	TaskID    uint      `gorm:"index" json:"task_id"`
	NoteType  string    `gorm:"default:'task'" json:"note_type"` // task or note
	Label     string    `json:"label"`
	Sort      int       `gorm:"default:0" json:"sort"`
	Content   string    `gorm:"not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
