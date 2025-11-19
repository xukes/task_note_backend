package models

import "time"

type Note struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	TaskID    string    `gorm:"index;not null" json:"task_id"`
	Content   string    `gorm:"not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
