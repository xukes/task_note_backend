package controllers

import (
	"net/http"
	"sync"
	"task_note_backend/database"
	"task_note_backend/models"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var tempTOTPSecrets sync.Map

type VerifyTOTPInput struct {
	Token string `json:"token" binding:"required"`
}

func GenerateTOTP(c *gin.Context) {
	userId, _ := c.Get("user_id")
	var user models.User
	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "TaskNote",
		AccountName: user.Username,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate TOTP key"})
		return
	}

	// Store secret in memory temporarily instead of database
	tempTOTPSecrets.Store(userId, key.Secret())

	c.JSON(http.StatusOK, gin.H{
		"secret": key.Secret(),
		"url":    key.URL(),
	})
}

func VerifyAndBindTOTP(c *gin.Context) {
	userId, _ := c.Get("user_id")
	var input VerifyTOTPInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if there is a pending secret in memory
	secretInterface, ok := tempTOTPSecrets.Load(userId)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No pending TOTP setup found or expired. Please generate a new QR code."})
		return
	}
	secret := secretInterface.(string)

	valid, _ := totp.ValidateCustom(input.Token, secret, totp.ValidateOpts{
		Period:    30,
		Skew:      2,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid TOTP token"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	// Save to database only after verification
	user.TOTPSecret = secret
	user.TOTPEnabled = true
	database.DB.Save(&user)

	// Clear temporary secret
	tempTOTPSecrets.Delete(userId)

	c.JSON(http.StatusOK, gin.H{"message": "TOTP enabled successfully"})
}

func GetTOTPStatus(c *gin.Context) {
	userId, _ := c.Get("user_id")
	var user models.User
	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled": user.TOTPEnabled,
	})
}
