package controllers

import (
	"fmt"
	"net/http"
	os "os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only image files are allowed"})
		return
	}

	// Ensure directory exists with correct permissions (755)
	uploadDir := "/front/build/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}
	// Force directory permissions to 755 in case they were messed up by previous code
	// MkdirAll doesn't change permissions if directory already exists
	if err := os.Chmod(uploadDir, 0755); err != nil {
		fmt.Printf("Failed to chmod directory: %v\n", err)
	}

	// Generate unique filename
	filename := uuid.New().String() + ext
	uploadPath := filepath.Join(uploadDir, filename)

	// Save file
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Change file permissions to 644 so nginx (and others) can read it
	if err := os.Chmod(uploadPath, 0644); err != nil {
		fmt.Printf("Failed to chmod file: %v\n", err)
		// Don't return error here, as file is saved
	}
	// Return public URL
	// Assuming server runs on localhost:8080. In production, this should be configured.
	url := fmt.Sprintf("http://8.152.101.46:8099/uploads/%s", filename)
	c.JSON(http.StatusOK, gin.H{"url": url})
}
