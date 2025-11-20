package controllers

import (
	"net/http"
	"sync"
	"time"
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

	// 将密钥暂时存储在内存中，而不是数据库
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

	// 检查内存中是否有待处理的密钥
	secretInterface, ok := tempTOTPSecrets.Load(userId)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到待处理的 TOTP 设置或已过期。请重新生成二维码。"})
		return
	}
	secret := secretInterface.(string)

	valid, _ := totp.ValidateCustom(input.Token, secret, time.Now(), totp.ValidateOpts{
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

	// 仅在验证通过后保存到数据库
	user.TOTPSecret = secret
	user.TOTPEnabled = true
	database.DB.Save(&user)

	// 清除临时密钥
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
