package models


type Task struct {
	ID          string `gorm:"primaryKey" json:"id"`
	UserID      uint   `gorm:"index;not null" json:"user_id"`
	Title       string `gorm:"not null" json:"title"`
	Completed   bool   `gorm:"default:false" json:"completed"`
	CreatedAt   int64  `json:"created_at"`           // Changed to int64 (timestamp in milliseconds)
	CompletedAt *int64 `json:"completed_at"`         // Pointer to allow null
	TimeSpent   int    `json:"time_spent" gorm:"default:0"` // Time spent in minutes
	TimeUnit    string `json:"time_unit" gorm:"default:'minute'"` // Unit: minute, hour, day, week, month
	Notes       []Note `json:"notes" gorm:"foreignKey:TaskID"`
}
