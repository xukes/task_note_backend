package search

import (
	"fmt"
	"log"
	"strconv"
	"task_note_backend/models"

	"github.com/blevesearch/bleve/v2"
)

var index bleve.Index
var noteIndex bleve.Index

func Init() {
	mapping := bleve.NewIndexMapping()
	var err error
	index, err = bleve.Open("task_index.bleve")
	if err == bleve.ErrorIndexPathDoesNotExist {
		index, err = bleve.New("task_index.bleve", mapping)
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}

	// Initialize Note Index
	noteIndex, err = bleve.Open("note_index.bleve")
	if err == bleve.ErrorIndexPathDoesNotExist {
		noteIndex, err = bleve.New("note_index.bleve", mapping)
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}
}

type TaskIndex struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	UserID  string `json:"user_id"`
}

type NoteIndex struct {
	ID      uint   `json:"id"`
	Content string `json:"content"`
	Label   string `json:"label"`
	UserID  string `json:"user_id"`
}

func IndexTask(task models.Task) {
	if index == nil {
		return
	}

	content := ""
	for _, note := range task.Notes {
		content += note.Content + " "
	}

	doc := TaskIndex{
		ID:      task.ID,
		Title:   task.Title,
		Content: content,
		UserID:  strconv.Itoa(int(task.UserID)),
	}

	err := index.Index(strconv.Itoa(int(task.ID)), doc)
	if err != nil {
		log.Printf("Error indexing task %d: %v", task.ID, err)
	}
}

func IndexNote(note models.Note) {
	if noteIndex == nil {
		return
	}

	doc := NoteIndex{
		ID:      note.ID,
		Content: note.Content,
		Label:   note.Label,
		UserID:  strconv.Itoa(int(note.UserID)),
	}

	err := noteIndex.Index(strconv.Itoa(int(note.ID)), doc)
	if err != nil {
		log.Printf("Error indexing note %d: %v", note.ID, err)
	}
}

func DeleteTask(taskId uint) {
	if index == nil {
		return
	}
	err := index.Delete(strconv.Itoa(int(taskId)))
	if err != nil {
		log.Printf("Error deleting task %d from index: %v", taskId, err)
	}
}

func DeleteNoteIndex(noteId uint) {
	if noteIndex == nil {
		return
	}
	err := noteIndex.Delete(strconv.Itoa(int(noteId)))
	if err != nil {
		log.Printf("Error deleting note %d from index: %v", noteId, err)
	}
}

func SearchTasks(queryStr string, userId uint) ([]SearchResult, error) {
	if index == nil {
		return nil, fmt.Errorf("index not initialized")
	}
	return searchIndex(index, queryStr, userId)
}

func SearchNotes(queryStr string, userId uint) ([]SearchResult, error) {
	if noteIndex == nil {
		return nil, fmt.Errorf("note index not initialized")
	}
	return searchIndex(noteIndex, queryStr, userId)
}

func searchIndex(idx bleve.Index, queryStr string, userId uint) ([]SearchResult, error) {
	// Filter by UserID
	userQuery := bleve.NewTermQuery(strconv.Itoa(int(userId)))
	userQuery.SetField("user_id")

	// The user's search text
	matchQuery := bleve.NewQueryStringQuery(queryStr)

	// Combine them
	conjunctionQuery := bleve.NewConjunctionQuery(userQuery, matchQuery)

	searchRequest := bleve.NewSearchRequest(conjunctionQuery)
	searchRequest.Size = 20 // Increased size
	searchRequest.Highlight = bleve.NewHighlight()

	searchResults, err := idx.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, hit := range searchResults.Hits {
		id, _ := strconv.Atoi(hit.ID)
		results = append(results, SearchResult{
			ID:        uint(id),
			Fragments: hit.Fragments,
		})
	}
	return results, nil
}

type SearchResult struct {
	ID        uint
	Fragments map[string][]string
}
