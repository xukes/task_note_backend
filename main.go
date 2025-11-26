package main

import (
	"task_note_backend/controllers"
	"task_note_backend/database"
	"task_note_backend/middleware"
	"task_note_backend/models"
	"task_note_backend/search"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDatabase()
	search.Init()

	// Seed default user
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		controllers.CreateUser("admin", "123456")
	}

	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	au := r.Group("/api")
	au.POST("/login", controllers.Login)
	au.POST("/register", controllers.Register)
	au.POST("/auth/reset-password", controllers.ResetPassword)

	// Serve static files from uploads directory
	au.Static("/uploads", "./uploads")

	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/upload", controllers.UploadFile)
		protected.GET("/search", controllers.SearchTasks)
		protected.GET("/search/notes", controllers.SearchNotes)

		protected.GET("/tasks", controllers.GetTasks)
		protected.GET("/tasks/stats", controllers.GetTaskStats)
		protected.POST("/tasks", controllers.CreateTask)
		protected.PUT("/tasks/:id", controllers.UpdateTask)
		protected.PATCH("/tasks/:id/toggle", controllers.ToggleTask)
		protected.DELETE("/tasks/:id", controllers.DeleteTask)

		protected.POST("/notes", controllers.CreateNote)
		protected.GET("/notes", controllers.GetNotes)
		protected.PUT("/notes/:id", controllers.UpdateNote)
		protected.DELETE("/notes/:id", controllers.DeleteNote)

		protected.POST("/auth/totp/generate", controllers.GenerateTOTP)
		protected.POST("/auth/totp/verify", controllers.VerifyAndBindTOTP)
		protected.GET("/auth/totp/status", controllers.GetTOTPStatus)
	}

	r.Run(":8080")
}
