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

	input.UserID = userId
	input.CreatedAt = time.Now()

	if input.NoteType == "note" {
		// Independent note
		input.TaskID = 0
	} else {
		// Task note (default)
		input.NoteType = "task"
		// Verify Task belongs to User
		var task models.Task
		if err := database.DB.Where("id = ? AND user_id = ?", input.TaskID, userId).First(&task).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found or access denied"})
			return
		}
	}

	if err := database.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if input.NoteType == "task" {
		// Re-index parent task asynchronously
		go func(taskId uint) {
			var parentTask models.Task
			database.DB.Preload("Notes").First(&parentTask, taskId)
			search.IndexTask(parentTask)
		}(input.TaskID)
	} else {
		// Index independent note
		go search.IndexNote(input)
	}

	c.JSON(http.StatusOK, input)
}

func GetNotes(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	noteType := c.Query("type")
	if noteType == "" {
		noteType = "note"
	}

	var notes []models.Note
	query := database.DB.Where("user_id = ?", userId)

	if noteType == "note" {
		query = query.Where("note_type = ?", "note")
	} else if noteType == "task" {
		query = query.Where("note_type = ?", "task")
	}

	if err := query.Order("sort desc, created_at desc").Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notes)
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

	// Check permission
	if note.NoteType == "note" {
		if note.UserID != userId {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	} else {
		// Task note
		var task models.Task
		if err := database.DB.First(&task, "id = ?", note.TaskID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if task.UserID != userId {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	note.Content = input.Content
	note.Label = input.Label
	note.Sort = input.Sort
	if err := database.DB.Save(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if note.NoteType == "task" {
		// Re-index parent task asynchronously
		go func(taskId uint) {
			var parentTask models.Task
			database.DB.Preload("Notes").First(&parentTask, taskId)
			search.IndexTask(parentTask)
		}(note.TaskID)
	} else {
		// Index independent note
		go search.IndexNote(note)
	}

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

	if note.NoteType == "note" {
		if note.UserID != userId {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	} else {
		var task models.Task
		if err := database.DB.First(&task, "id = ?", note.TaskID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if task.UserID != userId {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	if err := database.DB.Delete(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if note.NoteType == "task" {
		// Re-index parent task asynchronously
		go func(taskId uint) {
			var parentTask models.Task
			database.DB.Preload("Notes").First(&parentTask, taskId)
			search.IndexTask(parentTask)
		}(note.TaskID)
	} else {
		// Delete independent note index
		go search.DeleteNoteIndex(note.ID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted"})
}

func SearchNotes(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	queryStr := c.Query("q")

	if queryStr == "" {
		c.JSON(http.StatusOK, []models.Note{})
		return
	}

	results, err := search.SearchNotes(queryStr, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusOK, []models.Note{})
		return
	}

	var ids []uint
	for _, r := range results {
		ids = append(ids, r.ID)
	}

	var notes []models.Note
	if err := database.DB.Where("id IN ?", ids).Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reorder notes based on search result order
	noteMap := make(map[uint]models.Note)
	for _, n := range notes {
		noteMap[n.ID] = n
	}

	var orderedNotes []models.Note
	for _, id := range ids {
		if n, ok := noteMap[id]; ok {
			orderedNotes = append(orderedNotes, n)
		}
	}

	c.JSON(http.StatusOK, orderedNotes)
}
