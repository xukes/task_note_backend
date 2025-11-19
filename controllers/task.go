package controllers

import (
	"net/http"
	"task_note_backend/database"
	"task_note_backend/models"
	"time"

	"github.com/gin-gonic/gin"
)

func GetTasks(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	var tasks []models.Task

	query := database.DB.Preload("Notes").Where("user_id = ?", userId)

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr != "" && endDateStr != "" {
		// Assuming frontend sends timestamps as strings
		query = query.Where("created_at BETWEEN ? AND ?", startDateStr, endDateStr)
	}

	if err := query.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func CreateTask(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	var input models.Task
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.UserID = userId
	input.CreatedAt = time.Now().UnixMilli() // Save as milliseconds timestamp

	if err := database.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, input)
}

func UpdateTask(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	taskId := c.Param("id")

	var task models.Task
	if err := database.DB.Where("id = ? AND user_id = ?", taskId, userId).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var input models.Task
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle CompletedAt logic
	if input.Completed != task.Completed {
		if input.Completed {
			now := time.Now().UnixMilli()
			task.CompletedAt = &now
		} else {
			task.CompletedAt = nil
		}
	}

	// Only update allowed fields
	task.Title = input.Title
	task.Completed = input.Completed
	
	if err := database.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func DeleteTask(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	taskId := c.Param("id")

	var task models.Task
	if err := database.DB.Where("id = ? AND user_id = ?", taskId, userId).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if err := database.DB.Delete(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

func ToggleTask(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	taskId := c.Param("id")

	var task models.Task
	if err := database.DB.Where("id = ? AND user_id = ?", taskId, userId).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Toggle status
	task.Completed = !task.Completed
	if task.Completed {
		now := time.Now().UnixMilli()
		task.CompletedAt = &now
	} else {
		task.CompletedAt = nil
	}

	// Save changes
	if err := database.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}
