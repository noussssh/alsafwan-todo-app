package main

import (
	"embed"
	"log"
	"net/http"
	"os"

	"alsafwanmarine.com/todo-app/internal/database"
	"alsafwanmarine.com/todo-app/internal/handlers"
)

//go:embed web/static
var staticFiles embed.FS

func main() {
	// Initialize database
	db, err := database.InitDB("./data/todos.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Create handlers
	h := handlers.NewHandler(db)

	// Routes
	http.HandleFunc("/", h.HomeHandler)

	// Serve embedded static files
	staticFS, err := staticFiles.ReadDir("web/static")
	if err != nil {
		log.Fatal("Failed to read embedded static files:", err)
	}
	log.Printf("📁 Embedded static files: %d items", len(staticFS))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Hello Go App starting on http://localhost:%s", port)
	log.Printf("📦 Static files embedded in binary")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}