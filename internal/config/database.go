package config

import (
	"time"

	"alsafwanmarine.com/todo-app/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	// Configure SQLite with performance optimizations
	dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=temp_store(MEMORY)&_pragma=mmap_size(268435456)&_pragma=foreign_keys(ON)&_pragma=cache_size(-64000)"
	
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Reduce logging overhead in production
		PrepareStmt: true, // Enable prepared statement cache
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		return nil, err
	}
	
	// Get the underlying SQL database to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	
	// Configure connection pool for better performance
	sqlDB.SetMaxOpenConns(25)                 // Maximum number of open connections
	sqlDB.SetMaxIdleConns(25)                 // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection
	sqlDB.SetConnMaxIdleTime(time.Minute)     // Maximum idle time for a connection
	
	database := &Database{DB: db}
	
	if err := database.migrate(); err != nil {
		return nil, err
	}
	
	// Create indexes for better query performance
	if err := database.createIndexes(); err != nil {
		return nil, err
	}
	
	return database, nil
}

func (d *Database) migrate() error {
	return d.DB.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.UserActivity{},
		&models.PasswordResetEvent{},
	)
}

func (d *Database) createIndexes() error {
	// Performance-critical indexes for common queries
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email_enabled ON users(email, enabled);",
		"CREATE INDEX IF NOT EXISTS idx_users_role_enabled ON users(role, enabled);",
		"CREATE INDEX IF NOT EXISTS idx_users_enabled_created_at ON users(enabled, created_at DESC);",
		"CREATE INDEX IF NOT EXISTS idx_sessions_user_id_expires_at ON sessions(user_id, expires_at);",
		"CREATE INDEX IF NOT EXISTS idx_sessions_token_expires_at ON sessions(token, expires_at);",
		"CREATE INDEX IF NOT EXISTS idx_user_activities_user_id_performed_at ON user_activities(user_id, performed_at DESC);",
		"CREATE INDEX IF NOT EXISTS idx_user_activities_type_performed_at ON user_activities(activity_type, performed_at DESC);",
		"CREATE INDEX IF NOT EXISTS idx_user_activities_performed_at ON user_activities(performed_at DESC);",
		"CREATE INDEX IF NOT EXISTS idx_password_reset_events_user_id ON password_reset_events(user_id, created_at DESC);",
		"CREATE INDEX IF NOT EXISTS idx_password_reset_events_expires_at ON password_reset_events(expires_at);",
	}

	for _, index := range indexes {
		if err := d.DB.Exec(index).Error; err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) Seed() error {
	seedUsers := []models.User{
		{
			ID:      2,
			Email:   "sales4@alsafwanmarine.com",
			Name:    "Tom Charley",
			Role:    models.RoleSalesperson,
			Company: stringPtr("Al Safwan Marine"),
			Enabled: true,
		},
		{
			ID:      8,
			Email:   "manager@example.com",
			Name:    "Sales Manager",
			Role:    models.RoleManager,
			Company: stringPtr("Al Safwan Marine"),
			Enabled: true,
		},
		{
			ID:      9,
			Email:   "operations@alsafwanmarine.com",
			Name:    "Noushad Moidunny",
			Role:    models.RoleAdmin,
			Company: stringPtr("Al Safwan Marine"),
			Enabled: true,
		},
		{
			ID:      10,
			Email:   "sales1@alsafwanmarine.com",
			Name:    "Lubdha Vipin",
			Role:    models.RoleSalesperson,
			Company: stringPtr("Al Safwan Marine"),
			Enabled: true,
		},
		{
			ID:      11,
			Email:   "bd@alsafwanmarine.com",
			Name:    "Sandra Santosh",
			Role:    models.RoleSalesperson,
			Company: stringPtr("Al Safwan Marine"),
			Enabled: true,
		},
		{
			ID:      12,
			Email:   "marketing@alsafwanmarine.com",
			Name:    "Thomas Siby",
			Role:    models.RoleSalesperson,
			Company: stringPtr("Al Safwan Marine"),
			Enabled: true,
		},
		{
			ID:      13,
			Email:   "business@alsafwanmarine.com",
			Name:    "Silpa Chelathur",
			Role:    models.RoleSalesperson,
			Company: stringPtr("Al Safwan Marine"),
			Enabled: true,
		},
		{
			ID:      14,
			Email:   "sales5@alsafwanmarine.com",
			Name:    "Krishna Swaroop",
			Role:    models.RoleSalesperson,
			Company: stringPtr("Al Safwan Marine"),
			Enabled: true,
		},
		{
			ID:      18,
			Email:   "admin@example.com",
			Name:    "Admin User",
			Role:    models.RoleAdmin,
			Company: nil,
			Enabled: true,
		},
	}
	
	for _, user := range seedUsers {
		var existingUser models.User
		if err := d.DB.Where("id = ?", user.ID).First(&existingUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				defaultPassword := "password123"
				if err := user.SetPassword(defaultPassword); err != nil {
					return err
				}
				
				if err := d.DB.Create(&user).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}
	
	return nil
}

func stringPtr(s string) *string {
	return &s
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}