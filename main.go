package main

import (
	"log"
	"net/http"
	"os"

	"alsafwanmarine.com/todo-app/internal/database"
	"alsafwanmarine.com/todo-app/internal/handlers"
)

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
	http.HandleFunc("/api/todos", h.TodosHandler)
	http.HandleFunc("/api/todos/", h.TodoHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("🚀 Todo App starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
