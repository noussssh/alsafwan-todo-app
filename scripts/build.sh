#!/bin/bash
echo "🔨 Building Al Safwan Marine Todo App..."
go mod tidy
go build -o bin/todo-app main.go
echo "✅ Build complete: bin/todo-app"
