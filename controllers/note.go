package controllers

import (
	"net/http"
	"task_note_backend/database"
	"task_note_backend/models"
	"task_note_backend/search"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateNote(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	var input models.Note
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify Task belongs to User
	var task models.Task
	if err := database.DB.Where("id = ? AND user_id = ?", input.TaskID, userId).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found or access denied"})
		return
	}

	input.CreatedAt = time.Now()
	if err := database.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Re-index parent task asynchronously
	go func(taskId uint) {
		var parentTask models.Task
		database.DB.Preload("Notes").First(&parentTask, taskId)
		search.IndexTask(parentTask)
	}(input.TaskID)

	c.JSON(http.StatusOK, input)
}

func UpdateNote(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	noteId := c.Param("id")
	var input models.Note
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var note models.Note
	if err := database.DB.First(&note, "id = ?", noteId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	var task models.Task
	if err := database.DB.First(&task, "id = ?", note.TaskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.UserID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	note.Content = input.Content
	if err := database.DB.Save(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Re-index parent task asynchronously
	go func(taskId uint) {
		var parentTask models.Task
		database.DB.Preload("Notes").First(&parentTask, taskId)
		search.IndexTask(parentTask)
	}(note.TaskID)

	c.JSON(http.StatusOK, note)
}

func DeleteNote(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	noteId := c.Param("id")

	var note models.Note
	if err := database.DB.First(&note, "id = ?", noteId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	var task models.Task
	if err := database.DB.First(&task, "id = ?", note.TaskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if task.UserID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := database.DB.Delete(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Re-index parent task asynchronously
	go func(taskId uint) {
		var parentTask models.Task
		database.DB.Preload("Notes").First(&parentTask, taskId)
		search.IndexTask(parentTask)
	}(note.TaskID)

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted"})
}
