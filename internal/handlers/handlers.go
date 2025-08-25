package handlers

import (
    "database/sql"
    "html/template"
    "net/http"
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
    <title>Hello Go</title>
    <style>
        body {
            font-family: system-ui, -apple-system, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #ff7a18 0%, #af002d 50%, #ff6b35 100%);
            color: white;
        }
        .container {
            text-align: center;
            padding: 2rem;
            background: rgba(255, 255, 255, 0.1);
            border-radius: 20px;
            backdrop-filter: blur(10px);
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }
        h1 {
            font-size: 4rem;
            margin-bottom: 1rem;
            text-shadow: 0 4px 8px rgba(0,0,0,0.3);
            font-weight: 800;
        }
        p {
            font-size: 1.2rem;
            opacity: 0.9;
            margin: 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Hello Go! 🚀</h1>
        <p>Fresh start with SQLite & embedded assets ready</p>
    </div>
</body>
</html>
`

    t, _ := template.New("home").Parse(tmpl)
    t.Execute(w, nil)
}

