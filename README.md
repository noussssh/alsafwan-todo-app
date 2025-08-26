# ASM Tracker User Management System

A comprehensive user management and authentication system built in Go for the ASM Tracker (maritime company tracking application). This system provides robust authentication, authorization, user management, and activity tracking capabilities.

## Features

### Core Authentication & Authorization
- **User Authentication**: Email/password login with secure session management
- **Role-based Access Control**: Three roles (Admin, Manager, Salesperson) with granular permissions
- **Session Management**: Secure sessions with 30-minute auto-expiry and extension
- **Password Security**: BCrypt hashing, expiry after 30 days, strong password generation

### User Management
- **CRUD Operations**: Create, read, update, delete users with role-based permissions
- **User Status Management**: Enable/disable user accounts
- **Bulk Operations**: Bulk password resets and status changes
- **Profile Management**: Users can update their own profiles and change passwords

### Security Features
- **Rate Limiting**: 10 login attempts per 3 minutes per IP
- **CSRF Protection**: Token-based CSRF protection for state-changing operations
- **Input Validation & Sanitization**: Comprehensive validation with custom rules
- **Security Headers**: XSS protection, content type options, frame options, CSP
- **Activity Logging**: Comprehensive audit trail of all user actions

### Password Management
- **Password Reset System**: Token-based password resets with email integration
- **Automatic Resets**: Auto-reset expired passwords and inactive user passwords
- **Strong Password Generation**: 8-character passwords with mixed case, numbers, special chars
- **Password History**: Track all password reset events with reasons and metadata

### Activity Tracking
- **User Activities**: Login/logout, page views, password changes, user CRUD operations
- **Failed Login Tracking**: Track failed login attempts for security monitoring
- **Session Duration**: Calculate and track session durations
- **Audit Trail**: Complete audit trail with IP addresses, user agents, and metadata

## Technology Stack

- **Web Framework**: Gin (HTTP router and middleware)
- **ORM**: GORM (database abstraction and migrations)
- **Database**: SQLite (development) / PostgreSQL (production)
- **Password Hashing**: bcrypt
- **Session Management**: Database-backed sessions with secure tokens
- **Validation**: go-playground/validator with custom rules
- **Rate Limiting**: ulule/limiter
- **Testing**: Go standard testing + testify

## Database Schema

### Users Table
```sql
users (
  id INTEGER PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  password_digest TEXT NOT NULL,
  role INTEGER NOT NULL DEFAULT 2, -- 0=admin, 1=manager, 2=salesperson
  company TEXT,
  enabled BOOLEAN DEFAULT TRUE,
  last_sign_in_at DATETIME,
  current_sign_in_at DATETIME,
  sign_in_count INTEGER DEFAULT 0,
  password_reset_at DATETIME,
  password_expires_at DATETIME,
  managed_customers_count INTEGER DEFAULT 0,
  created_at DATETIME,
  updated_at DATETIME
)
```

### Sessions Table
```sql
sessions (
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL,
  token TEXT UNIQUE NOT NULL,
  ip_address TEXT,
  user_agent TEXT,
  expires_at DATETIME NOT NULL,
  created_at DATETIME,
  updated_at DATETIME
)
```

### User Activities Table
```sql
user_activities (
  id INTEGER PRIMARY KEY,
  user_id INTEGER,
  activity_type TEXT NOT NULL,
  subject_type TEXT,
  subject_id INTEGER,
  ip_address TEXT,
  user_agent TEXT,
  session_duration INTEGER,
  metadata JSON,
  performed_at DATETIME NOT NULL
)
```

### Password Reset Events Table
```sql
password_reset_events (
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL,
  admin_id INTEGER,
  reason TEXT,
  ip_address TEXT,
  user_agent TEXT,
  success BOOLEAN DEFAULT FALSE,
  reset_type TEXT NOT NULL, -- manual, automatic_expiry, automatic_inactivity
  token TEXT UNIQUE,
  expires_at DATETIME,
  created_at DATETIME
)
```

## API Endpoints

### Authentication Endpoints
```
POST   /api/v1/auth/login              - User login
DELETE /api/v1/auth/logout             - User logout
GET    /api/v1/auth/me                 - Get current user info
PATCH  /api/v1/auth/password           - Change password
```

### Password Reset Endpoints
```
POST   /api/v1/passwords               - Request password reset
PATCH  /api/v1/passwords/reset         - Reset password with token
```

### User Management Endpoints (Admin/Manager)
```
GET    /api/v1/users                   - List users
GET    /api/v1/users/:id               - Get user details
POST   /api/v1/users                   - Create new user
PATCH  /api/v1/users/:id               - Update user
DELETE /api/v1/users/:id               - Delete user
POST   /api/v1/users/:id/reset_password - Reset user password
PATCH  /api/v1/users/:id/toggle_enabled - Enable/disable user
POST   /api/v1/users/bulk_reset_passwords - Bulk reset passwords
POST   /api/v1/users/bulk_toggle_enabled  - Bulk enable/disable users
GET    /api/v1/users/password_reset_events - View password reset history
```

### Activity Endpoints
```
GET    /api/v1/activities              - Get all activities (Admin/Manager)
GET    /api/v1/activities/users/:id    - Get user activities
```

### Health Check
```
GET    /health                         - Health check endpoint
```

## Installation & Setup

### Prerequisites
- Go 1.21 or higher
- SQLite (development) or PostgreSQL (production)

### Installation
1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd alsafwan-todo-app
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Run database migrations and seed data:
   ```bash
   go run main.go
   ```

### Environment Variables
```env
DATABASE_URL=data/asm_tracker.db  # SQLite path or PostgreSQL URL
PORT=8080                         # Server port
GIN_MODE=release                  # Gin mode (debug/release)
```

### Default Users
The system seeds the following users on first run:
- **Admin**: admin@example.com (admin role)
- **Operations**: operations@alsafwanmarine.com (admin role)
- **Manager**: manager@example.com (manager role)
- **Salespeople**: Various sales emails (salesperson role)

All seeded users have the default password: `password123`

## Usage Examples

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "password123"}'
```

### Create User (Admin/Manager)
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Cookie: session_token=<token>" \
  -d '{
    "email": "newuser@example.com",
    "name": "New User",
    "role": 2,
    "company": "Al Safwan Marine",
    "password": "strongpassword123"
  }'
```

### Reset Password
```bash
curl -X POST http://localhost:8080/api/v1/users/123/reset_password \
  -H "Content-Type: application/json" \
  -H "Cookie: session_token=<token>" \
  -d '{"reason": "User requested password reset"}'
```

## Role-Based Permissions

### Admin (Role 0)
- Full system access
- Can manage all users (create, update, delete)
- Can reset any user's password
- Can enable/disable salespeople
- Can view all activities and audit logs

### Manager (Role 1)
- Can manage salespeople only (create, update, delete)
- Cannot change user roles
- Can reset salesperson passwords
- Can enable/disable salespeople
- Can view salesperson activities

### Salesperson (Role 2)
- Can view own profile and change own password
- Can view assigned customers (when customer system is integrated)
- Cannot manage other users

## Security Considerations

### Authentication Security
- Passwords are hashed using bcrypt with cost factor 12+
- Sessions expire after 30 minutes of inactivity
- Rate limiting prevents brute force attacks
- Failed login attempts are logged and monitored

### Authorization Security
- Role-based access control with principle of least privilege
- Users cannot escalate their own privileges
- Managers cannot disable admins or other managers
- Self-service operations (like self-disable) are prevented

### Data Security
- Input validation and sanitization on all endpoints
- CSRF protection for state-changing operations
- SQL injection prevention through ORM
- XSS protection through security headers
- Secure cookie settings (HttpOnly, Secure, SameSite)

### Audit Security
- Comprehensive activity logging
- Password reset events tracking
- IP address and user agent logging
- Failed login attempt monitoring

## Testing

Run the test suite:
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test package
go test ./internal/models
go test ./internal/services
```

### Test Coverage Areas
- User model validation and business logic
- Authentication flow (login, logout, session management)
- Authorization middleware and role checking
- Password reset functionality
- Activity logging
- Input validation and sanitization

## Development

### Project Structure
```
alsafwan-todo-app/
├── cmd/                    # Command-line applications
├── internal/
│   ├── app/               # Application setup and routing
│   ├── config/            # Configuration and database setup
│   ├── controllers/       # HTTP handlers
│   ├── middleware/        # HTTP middleware (auth, CSRF, etc.)
│   ├── models/           # Data models and business logic
│   ├── services/         # Business logic services
│   └── utils/            # Utility functions
├── migrations/           # Database migrations
├── data/                # Database files (SQLite)
├── scripts/             # Build and deployment scripts
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── main.go              # Application entry point
└── README.md            # This file
```

### Adding New Features
1. Create models in `internal/models/`
2. Add business logic in `internal/services/`
3. Create HTTP handlers in `internal/controllers/`
4. Add routes in `internal/app/app.go`
5. Write tests for all components

### Database Migrations
The system uses GORM's AutoMigrate for development. For production:
1. Create migration files in `migrations/`
2. Use proper migration tools for schema changes
3. Always backup data before migrations

## Deployment

### Production Deployment
1. Build the binary:
   ```bash
   go build -o asm-tracker main.go
   ```

2. Set production environment variables:
   ```bash
   export GIN_MODE=release
   export DATABASE_URL=postgres://user:pass@host:port/dbname
   export PORT=8080
   ```

3. Run the application:
   ```bash
   ./asm-tracker
   ```

### Docker Deployment
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o asm-tracker main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/asm-tracker .
COPY --from=builder /app/data ./data
CMD ["./asm-tracker"]
```

## Migration from Rails

This Go implementation is designed to be compatible with existing Rails applications:

### Database Compatibility
- Compatible bcrypt password hashes
- Same user roles and permissions structure
- Matching session cookie format (configurable)
- Database schema compatibility

### API Compatibility
- RESTful API endpoints match Rails conventions
- JSON response formats compatible with existing frontend
- Authentication flow compatible with existing clients

### Migration Strategy
1. **Parallel Deployment**: Run both systems simultaneously
2. **Gradual Migration**: Move endpoints one by one
3. **Data Synchronization**: Keep databases in sync during transition
4. **Complete Cutover**: Switch traffic entirely to Go system

## Contributing

### Code Style
- Follow Go conventions and best practices
- Use `gofmt` for code formatting
- Add comments for exported functions and types
- Write tests for all new functionality

### Pull Request Process
1. Create feature branch from main
2. Write code following project conventions
3. Add comprehensive tests
4. Update documentation as needed
5. Submit pull request with detailed description

### Issue Reporting
- Use GitHub issues for bug reports and feature requests
- Include detailed reproduction steps for bugs
- Provide context and use cases for feature requests

## License

This project is proprietary software for Al Safwan Marine. All rights reserved.

## Support

For technical support and questions:
- Email: operations@alsafwanmarine.com
- Internal documentation: [Company Wiki]
- Issue tracking: GitHub Issues

---

**ASM Tracker User Management System** - Built with ❤️ for Al Safwan Marine