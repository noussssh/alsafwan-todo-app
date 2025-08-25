#!/bin/bash
# Deploy to Linux server
echo "🚀 Deploying to Al Safwan Marine Linux Server..."

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/todo-app-linux main.go

# Copy to server (adjust server details)
# scp bin/todo-app-linux alsafwan@your-server:/home/alsafwan/alsafwanmarine-project/applications/site1/
# ssh alsafwan@your-server "cd /home/alsafwan/alsafwanmarine-project/applications/site1 && ./todo-app-linux"

echo "✅ Built for Linux: bin/todo-app-linux"
echo "📦 Ready for deployment to server"
