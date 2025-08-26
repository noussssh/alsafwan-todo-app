package main

import (
	"log"
	"os"

	"alsafwanmarine.com/todo-app/internal/app"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found")
	}

	// Ensure required directories exist
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Printf("Warning: Could not create data directory: %v", err)
	}
	if err := os.MkdirAll("static", 0755); err != nil {
		log.Printf("Warning: Could not create static directory: %v", err)
	}
	if err := os.MkdirAll("templates", 0755); err != nil {
		log.Printf("Warning: Could not create templates directory: %v", err)
	}

	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		dbPath = "data/asm_tracker.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	// Set Gin mode for production
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	log.Printf("Starting ASM Tracker...")
	log.Printf("Database path: %s", dbPath)
	log.Printf("Port: %s", port)
	
	application, err := app.New(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}
	defer application.Close()

	r := gin.Default()
	application.SetupRoutes(r)

	log.Printf("ASM Tracker User Management System starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
