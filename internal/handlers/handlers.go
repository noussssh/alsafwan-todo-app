package handlers

import (
    "database/sql"
    "encoding/json"
    "html/template"
    "net/http"
    "strconv"
    "strings"

    "alsafwanmarine.com/todo-app/internal/database"
)

type Handler struct {
    db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
    return &Handler{db: db}
}

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
    tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Al Safwan Marine - Todo App</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <div class="container">
        <header>
            <h1>🚢 Al Safwan Marine Todo App</h1>
            <p>Development Environment - MacBook to Linux Pipeline</p>
        </header>

        <div class="todo-form">
            <h3>Add New Todo</h3>
            <form id="todoForm">
                <input type="text" id="title" placeholder="Todo title" required>
                <textarea id="description" placeholder="Description (optional)" rows="3"></textarea>
                <button type="submit">Add Todo</button>
            </form>
        </div>

        <div class="todos-container">
            <h3>Your Todos</h3>
            <div id="todosList"></div>
        </div>
    </div>

    <script src="/static/js/app.js"></script>
</body>
</html>
`

    t, _ := template.New("home").Parse(tmpl)
    t.Execute(w, nil)
}

func (h *Handler) TodosHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    switch r.Method {
    case "GET":
        todos, err := database.GetAllTodos(h.db)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(todos)

    case "POST":
        var req struct {
            Title       string `json:"title"`
            Description string `json:"description"`
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        todo, err := database.CreateTodo(h.db, req.Title, req.Description)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(todo)
    }
}

func (h *Handler) TodoHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // Extract ID from path
    path := strings.TrimPrefix(r.URL.Path, "/api/todos/")
    id, err := strconv.Atoi(path)
    if err != nil {
        http.Error(w, "Invalid todo ID", http.StatusBadRequest)
        return
    }

    switch r.Method {
    case "PUT":
        var req struct {
            Title       string `json:"title"`
            Description string `json:"description"`
            Completed   bool   `json:"completed"`
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        todo, err := database.UpdateTodo(h.db, id, req.Title, req.Description, req.Completed)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        json.NewEncoder(w).Encode(todo)

    case "DELETE":
        if err := database.DeleteTodo(h.db, id); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusNoContent)
    }
}
