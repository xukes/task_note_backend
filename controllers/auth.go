package controllers

import (
	"net/http"
	"task_note_backend/database"
	"task_note_backend/models"
	"task_note_backend/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	TOTPToken string `json:"totp_token"`
}

type ResetPasswordInput struct {
	Username    string `json:"username" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
	TOTPToken   string `json:"totp_token" binding:"required"`
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}
	user.TOTPEnabled = false
	// 检查是否启用了双重认证 (2FA)
	if user.TOTPEnabled {
		if input.TOTPToken == "" {
			c.JSON(http.StatusOK, gin.H{
				"require_2fa": true,
				"token":       "",
				"username":    "",
				"message":     "Please provide 2FA token",
			})
			return
		}

		valid, _ := totp.ValidateCustom(input.TOTPToken, user.TOTPSecret, time.Now(), totp.ValidateOpts{
			Period:    30,
			Skew:      2,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		})
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 2FA token"})
			return
		}
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"require_2fa": false,
		"token":       token,
		"username":    user.Username,
		"message":     "Login successful",
	})
}

func Register(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var existingUser models.User
	if err := database.DB.Where("username = ?", input.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{Username: input.Username, Password: string(hashedPassword)}
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}

// Helper to seed user if needed
func CreateUser(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := models.User{Username: username, Password: string(hashedPassword)}
	return database.DB.Create(&user).Error
}

func ResetPassword(c *gin.Context) {
	var input ResetPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !user.TOTPEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA is not enabled for this account. Cannot reset password via TOTP."})
		return
	}

	valid, _ := totp.ValidateCustom(input.TOTPToken, user.TOTPSecret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      2,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 2FA token"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user.Password = string(hashedPassword)
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}
