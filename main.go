package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/krushnna/meeting-scheduler/initializers"
	"github.com/krushnna/meeting-scheduler/routers"
	"github.com/krushnna/meeting-scheduler/utils"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}
	// Initialize logger
	utils.InitLogger()
	defer utils.Logger.Sync()

	logger := utils.GetLogger()

	// Initialize the databases and auto-migrate models
	db := initializers.InitDB()

	// Set up the router
	router := routers.SetupRouter(db, logger)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	log.Printf("Server starting on port %s......", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server:- %v", err)
	}
}
