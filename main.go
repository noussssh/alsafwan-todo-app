package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"alsafwanmarine.com/todo-app/internal/app"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

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
	log.Printf("Working directory: %s", func() string { wd, _ := os.Getwd(); return wd }())
	
	// Check if we can write to the data directory
	testFile := filepath.Join("data", "test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		log.Printf("Warning: Cannot write to data directory: %v", err)
	} else {
		os.Remove(testFile)
		log.Printf("Data directory is writable")
	}
	
	application, err := app.New(dbPath, templatesFS, staticFS)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.Close()

	r := gin.Default()
	
	// Add a simple test endpoint that always works
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})
	
	application.SetupRoutes(r)

	log.Printf("ASM Tracker routes configured successfully")
	log.Printf("ASM Tracker User Management System starting on port %s", port)
	
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server on port %s: %v", port, err)
	}
}
