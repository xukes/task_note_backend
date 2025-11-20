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
	tasks := []models.Task{}

	query := database.DB.Preload("Notes").Where("user_id = ?", userId)

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr != "" && endDateStr != "" {
		// Filter by task_time instead of created_at
		query = query.Where("task_time BETWEEN ? AND ?", startDateStr, endDateStr)
	}

	if err := query.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

type DailyTaskStat struct {
	Date             string `json:"date"`
	TotalCount       int    `json:"total_count"`
	UnCompletedCount int    `json:"un_completed_count"`
}

func GetTaskStats(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	stats := []DailyTaskStat{}

	query := `
		SELECT
		  date(task_time/1000, 'unixepoch', '+8 hours') AS date,
		  COUNT(*) AS total_count,
		  SUM(CASE WHEN completed = 0 THEN 1 ELSE 0 END) AS un_completed_count
		FROM tasks
		WHERE user_id = ? AND task_time BETWEEN ? AND ?
		GROUP BY date
		ORDER BY date
	`

	if err := database.DB.Raw(query, userId, startDateStr, endDateStr).Scan(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
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
	// If TaskTime wasn't provided, default it to CreatedAt for backward compatibility
	if input.TaskTime == 0 {
		input.TaskTime = input.CreatedAt
	}

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

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})

	// Update Title if present
	if title, ok := input["title"].(string); ok {
		updates["title"] = title
		task.Title = title // Update struct for response
	}

	// Update Completed if present
	if completed, ok := input["completed"].(bool); ok {
		if completed != task.Completed {
			updates["completed"] = completed
			task.Completed = completed // Update struct for response

			if completed {
				now := time.Now().UnixMilli()
				updates["completed_at"] = now
				task.CompletedAt = &now // Update struct for response
			} else {
				updates["completed_at"] = nil
				task.CompletedAt = nil // Update struct for response
			}
		}
	}

	// Update TimeSpent if present
	if timeSpent, ok := input["time_spent"].(float64); ok {
		updates["time_spent"] = int(timeSpent)
		task.TimeSpent = int(timeSpent) // Update struct for response
	}

	// Update TimeUnit if present
	if timeUnit, ok := input["time_unit"].(string); ok {
		updates["time_unit"] = timeUnit
		task.TimeUnit = timeUnit // Update struct for response
	}

	// Update TaskTime if present
	if taskTime, ok := input["task_time"].(float64); ok {
		updates["task_time"] = int64(taskTime)
		task.TaskTime = int64(taskTime)
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&task).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
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
