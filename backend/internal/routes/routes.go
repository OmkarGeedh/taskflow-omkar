package routes

import (
	"github.com/gin-gonic/gin"

	"taskflow/backend/internal/controllers"
	"taskflow/backend/internal/middleware"
)

func SetupRoutes(router *gin.Engine) {
	// Auth routes will be public
	auth := router.Group("/auth")
	{
		auth.POST("/register", controllers.Register())
		auth.POST("/login", controllers.Login())
	}

	// Protected routes — JWT required
	api := router.Group("/")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/profile", controllers.GetProfile())

		// Project routes
		api.GET("/projects", controllers.ListProjects())
		api.POST("/projects", controllers.CreateProject())
		api.GET("/projects/:id", controllers.GetProject())
		api.PUT("/projects/:id", controllers.UpdateProject())
		api.DELETE("/projects/:id", controllers.DeleteProject())

		// Task routes
		api.GET("/projects/:id/tasks", controllers.ListTasks())
		api.POST("/projects/:id/tasks", controllers.CreateTask())
		api.PATCH("/tasks/:id", controllers.UpdateTask())
		api.DELETE("/tasks/:id", controllers.DeleteTask())
	}
}
