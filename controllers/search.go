package controllers

import (
	"net/http"
	"task_note_backend/database"
	"task_note_backend/models"
	"task_note_backend/search"

	"github.com/gin-gonic/gin"
)

func SearchTasks(c *gin.Context) {
	userId := c.MustGet("user_id").(uint)
	queryStr := c.Query("q")

	if queryStr == "" {
		c.JSON(http.StatusOK, []models.Task{})
		return
	}

	results, err := search.SearchTasks(queryStr, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusOK, []models.Task{})
		return
	}

	var ids []uint
	highlightsMap := make(map[uint]map[string][]string)
	for _, r := range results {
		ids = append(ids, r.ID)
		highlightsMap[r.ID] = r.Fragments
	}

	var tasks []models.Task
	// Fetch tasks from DB. Note: WHERE IN does not guarantee order.
	// We need to reorder them based on the search result order.
	if err := database.DB.Preload("Notes").Where("id IN ?", ids).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Reorder tasks
	taskMap := make(map[uint]models.Task)
	for _, t := range tasks {
		for j := range t.Notes {
			t.Notes[j].Content = processNoteContent(t.Notes[j].Content, c)
		}
		taskMap[t.ID] = t
	}

	type TaskResponse struct {
		models.Task
		Highlights map[string][]string `json:"highlights"`
	}

	orderedTasks := make([]TaskResponse, 0, len(ids))
	for _, id := range ids {
		if t, ok := taskMap[id]; ok {
			orderedTasks = append(orderedTasks, TaskResponse{
				Task:       t,
				Highlights: highlightsMap[id],
			})
		}
	}

	c.JSON(http.StatusOK, orderedTasks)
}
