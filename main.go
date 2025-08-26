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

	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		dbPath = "data/asm_tracker.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	application, err := app.New(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}
	defer application.Close()

	r := gin.Default()
	application.SetupRoutes(r)

	log.Printf("ASM Tracker User Management System starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
