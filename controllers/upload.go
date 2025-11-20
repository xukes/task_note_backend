package controllers

import (
	"fmt"
	"io"
	"net/http"
	os "os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
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

	// Generate unique filename base
	filenameBase := uuid.New().String()
	var filename string
	var uploadPath string

	// Check if file size > 500KB (500 * 1024 bytes)
	if file.Size > 500*1024 {
		// Open the uploaded file
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
			return
		}
		defer src.Close()

		// Read file content
		buffer, err := io.ReadAll(src)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
			return
		}

		// Convert to WebP using bimg
		newImage, err := bimg.NewImage(buffer).Convert(bimg.WEBP)
		if err != nil {
			fmt.Printf("Image compression failed: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compress image"})
			return
		}

		filename = filenameBase + ".webp"
		uploadPath = filepath.Join(uploadDir, filename)

		// Save compressed file
		if err := bimg.Write(uploadPath, newImage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save compressed file"})
			return
		}
	} else {
		// Save original file
		filename = filenameBase + ext
		uploadPath = filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, uploadPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
	}

	// Change file permissions to 644 so nginx (and others) can read it
	if err := os.Chmod(uploadPath, 0644); err != nil {
		fmt.Printf("Failed to chmod file: %v\n", err)
		// Don't return error here, as file is saved
	}
	// Force directory permissions to 755 in case they were messed up by previous code
	// MkdirAll doesn't change permissions if directory already exists
	if err := os.Chmod(uploadDir, 0755); err != nil {
		fmt.Printf("Failed to chmod directory: %v\n", err)
	}

	// Return public URL
	// Assuming server runs on localhost:8080. In production, this should be configured.
	url := fmt.Sprintf("http://8.152.101.46:8099/uploads/%s", filename)
	c.JSON(http.StatusOK, gin.H{"url": url})
}
