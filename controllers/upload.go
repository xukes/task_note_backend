package controllers

import (
	"fmt"
	"net/http"
	os "os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
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
	uploadDir := "./uploads"
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

		// Decode image
		img, err := imaging.Decode(src)
		if err != nil {
			// If decoding fails, fall back to saving original
			fmt.Printf("Image decode failed (saving original): %v\n", err)
			filename = filenameBase + ext
			uploadPath = filepath.Join(uploadDir, filename)
			if err := c.SaveUploadedFile(file, uploadPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
				return
			}
		} else {
			// Compress: Resize if too large (optional)
			// Limit width to 1920px
			if img.Bounds().Dx() > 1920 {
				img = imaging.Resize(img, 1920, 0, imaging.Lanczos)
			}

			// Save as WebP
			filename = filenameBase + ".webp"
			uploadPath = filepath.Join(uploadDir, filename)

			// Create output file
			out, err := os.Create(uploadPath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create output file"})
				return
			}
			defer out.Close()

			// Encode as WebP with quality 75
			err = webp.Encode(out, img, &webp.Options{Lossless: false, Quality: 75})
			if err != nil {
				fmt.Printf("WebP encode failed: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save compressed file"})
				return
			}
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
	}
	// Force directory permissions to 755
	if err := os.Chmod(uploadDir, 0755); err != nil {
		fmt.Printf("Failed to chmod directory: %v\n", err)
	}

	// Return public URL
	url := fmt.Sprintf("http://8.152.101.46:8099/uploads/%s", filename)
	c.JSON(http.StatusOK, gin.H{"url": url})
}
