package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/omkargeedh/taskflow/internal/config"
	"github.com/omkargeedh/taskflow/internal/database"
	"github.com/omkargeedh/taskflow/internal/routes"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
