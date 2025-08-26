package config

import (
	"alsafwanmarine.com/todo-app/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	
	database := &Database{DB: db}
	
	if err := database.migrate(); err != nil {
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