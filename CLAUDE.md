# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Running the Application
- **Development**: `./scripts/dev.sh` or `go run main.go` - Starts server on port 8001 (configurable via PORT env var)
- **Build**: `./scripts/build.sh` or `go build -o bin/todo-app main.go`
- **Deploy**: `./scripts/deploy.sh` (production deployment)

### Testing
- **Run all tests**: `go test ./...`
- **Test with coverage**: `go test -cover ./...`
- **Test specific package**: `go test ./internal/models` or `./internal/services`

### Database
- **Database path**: `data/asm_tracker.db` (SQLite, configurable via DATABASE_URL)
- **Auto-migration**: Handled by GORM on startup
- **Seeded users**: admin@example.com, operations@alsafwanmarine.com, manager@example.com (password: `password123`)

## Architecture Overview

This is a Go web application using Gin framework for an ASM Tracker User Management System with embedded templates and static files.

### Core Architecture Pattern
- **MVC Structure**: Controllers handle HTTP requests, Services contain business logic, Models define data structures
- **Dependency Injection**: Application struct (`internal/app/app.go`) wires all dependencies together
- **Embedded Resources**: Templates and static files are embedded using `//go:embed` for deployment

### Key Components
- **Authentication**: Session-based with bcrypt password hashing, role-based access (Admin/Manager/Salesperson)
- **Database**: GORM with SQLite (dev) / PostgreSQL (prod), auto-migrations
- **Security**: Rate limiting, CSRF protection, input sanitization, security headers
- **Background Tasks**: Automatic session cleanup and password expiry handling

### Directory Structure
```
internal/
├── app/           # Application setup and dependency wiring
├── config/        # Database configuration and connection
├── controllers/   # HTTP handlers (web_*.go for web interface)
├── middleware/    # HTTP middleware (auth, security, validation)
├── models/        # Data models and business logic
├── services/      # Business logic services
└── utils/         # Shared utilities
```

### Service Layer Architecture
- **AuthService**: Login/logout, password changes, session management
- **SessionService**: Session CRUD operations and cleanup
- **ActivityService**: User activity logging and audit trails
- **PasswordResetService**: Password resets with token-based flow

### Database Models
- **Users**: Role-based (0=admin, 1=manager, 2=salesperson) with password expiry
- **Sessions**: Database-backed sessions with 30-minute expiry
- **UserActivities**: Comprehensive activity logging
- **PasswordResetEvents**: Password reset audit trail

### Web Interface
- **Template System**: Embedded HTML templates with base layout and partials
- **Static Assets**: Embedded CSS/JS served via `/static/` route
- **Flash Messages**: Session-based messaging system
- **CSRF Protection**: Token-based protection for state-changing operations

### Security Implementation
- **Rate Limiting**: 10 login attempts per 3 minutes per IP
- **Role Permissions**: Hierarchical (Admin > Manager > Salesperson)
- **Session Security**: HTTPOnly cookies, 30-minute expiry with extension
- **Input Validation**: Custom validation middleware with sanitization

### Port Configuration
- **Default Port**: 8001 (changeable via PORT environment variable)
- **Port Conflicts**: Use `lsof -ti:8001 | xargs kill -9` to free the port if needed

## Development Notes

### Adding New Features
1. Create models in `internal/models/` with GORM tags
2. Add business logic in `internal/services/`
3. Create controllers in `internal/controllers/`
4. Add routes in `internal/app/app.go`
5. Write tests alongside implementation

### Authentication Flow
- Login creates database session with token
- Middleware validates session token on each request
- Activity logging happens automatically via middleware
- Password expiry and session cleanup run hourly in background

### Role-Based Authorization
- Use `middleware.RequireWebRole(models.RoleManager)` for role restrictions
- Managers can only manage salespeople, not other managers/admins
- Self-service operations (profile updates) bypass role restrictions

### Template Development
- Templates are embedded at compile time from `templates/` directory
- Use base layout pattern: `layouts/base.html` with `partials/` for reusable components
- Flash messages available in all templates via middleware

### Database Development
- GORM handles auto-migrations on startup
- Seed data created automatically on first run
- Use `database.DB` directly for complex queries, services for business logic