package database

import (
	"log"
	"task_note_backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(sqlite.Open("tasks.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database!", err)
	}

	// Enable WAL mode for better concurrency
	database.Exec("PRAGMA journal_mode=WAL;")

	err = database.AutoMigrate(&models.User{}, &models.Task{}, &models.Note{})
	if err != nil {
		log.Fatal("Failed to migrate database!", err)
	}

	DB = database
}
