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
</head>
<body>
    <h1>Hello Go!</h1>
</body>
</html>
`

    t, _ := template.New("home").Parse(tmpl)
    t.Execute(w, nil)
}

