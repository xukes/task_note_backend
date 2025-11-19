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

	// Generate unique filename
	filename := uuid.New().String() + ext
	uploadPath := filepath.Join("/front/build/uploads", filename)

	// Save file
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	_ = os.Chmod(uploadPath, 0644)
	// Return public URL
	// Assuming server runs on localhost:8080. In production, this should be configured.
	url := fmt.Sprintf("http://8.152.101.46:8099/uploads/%s", filename)
	c.JSON(http.StatusOK, gin.H{"url": url})
}
