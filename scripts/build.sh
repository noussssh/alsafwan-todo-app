#!/bin/bash
echo "🔨 Building Al Safwan Marine Todo App..."

# Clean and optimize dependencies
go mod tidy
go mod download

# Create output directory
mkdir -p bin

# Build with optimizations for production
echo "🚀 Building with performance optimizations..."
CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.Version=$(git describe --tags --always --dirty) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -trimpath \
    -o bin/todo-app \
    main.go

# Verify binary size
echo "📊 Binary size: $(du -h bin/todo-app | cut -f1)"

# Optional: Strip additional symbols for even smaller binary (Linux/Unix only)
if command -v strip &> /dev/null; then
    echo "🔧 Stripping additional symbols..."
    strip bin/todo-app
    echo "📊 Stripped binary size: $(du -h bin/todo-app | cut -f1)"
fi

echo "✅ Build complete: bin/todo-app"
echo "💡 Performance tip: Set GIN_MODE=release and optimize database path for production"
