package database

import (
    "database/sql"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

type Todo struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

func InitDB(dbPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
    if err != nil {
        return nil, err
    }

    if err := createTables(db); err != nil {
        return nil, err
    }

    return db, nil
}

func createTables(db *sql.DB) error {
    schema := `
    CREATE TABLE IF NOT EXISTS todos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title VARCHAR(200) NOT NULL,
        description TEXT,
        completed BOOLEAN DEFAULT FALSE,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
    CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos(created_at);

    -- Enable WAL mode for better concurrency
    PRAGMA journal_mode = WAL;
    PRAGMA synchronous = NORMAL;
    PRAGMA cache_size = 1000;
    PRAGMA foreign_keys = ON;
    PRAGMA temp_store = memory;
    `

    _, err := db.Exec(schema)
    return err
}

func GetAllTodos(db *sql.DB) ([]Todo, error) {
    rows, err := db.Query(`
        SELECT id, title, description, completed, created_at, updated_at
        FROM todos
        ORDER BY created_at DESC
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var todos []Todo
    for rows.Next() {
        var todo Todo
        err := rows.Scan(&todo.ID, &todo.Title, &todo.Description,
            &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
        if err != nil {
            return nil, err
        }
        todos = append(todos, todo)
    }

    return todos, nil
}

func CreateTodo(db *sql.DB, title, description string) (*Todo, error) {
    result, err := db.Exec(`
        INSERT INTO todos (title, description)
        VALUES (?, ?)
    `, title, description)
    if err != nil {
        return nil, err
    }

    id, err := result.LastInsertId()
    if err != nil {
        return nil, err
    }

    return GetTodoByID(db, int(id))
}

func GetTodoByID(db *sql.DB, id int) (*Todo, error) {
    var todo Todo
    err := db.QueryRow(`
        SELECT id, title, description, completed, created_at, updated_at
        FROM todos WHERE id = ?
    `, id).Scan(&todo.ID, &todo.Title, &todo.Description,
        &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
    if err != nil {
        return nil, err
    }
    return &todo, nil
}

func UpdateTodo(db *sql.DB, id int, title, description string, completed bool) (*Todo, error) {
    _, err := db.Exec(`
        UPDATE todos
        SET title = ?, description = ?, completed = ?, updated_at = CURRENT_TIMESTAMP
        WHERE id = ?
    `, title, description, completed, id)
    if err != nil {
        return nil, err
    }

    return GetTodoByID(db, id)
}

func DeleteTodo(db *sql.DB, id int) error {
    _, err := db.Exec("DELETE FROM todos WHERE id = ?", id)
    return err
}
