package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"taskflow/backend/internal/config"
	"taskflow/backend/internal/database"
	"taskflow/backend/internal/routes"
)

func main() {
	// Load configuration from environment
	if err := config.Load(); err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Initialise database
	if err := database.InitDatabase(config.AppConfig.DatabaseURL); err != nil {
		log.Fatalf("failed to initialise database: %v", err)
	}
	defer database.CloseDatabase()

	// Setup routes
	router := gin.Default()
	routes.SetupRoutes(router)

	port := config.AppConfig.Port
	log.Printf("server listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
