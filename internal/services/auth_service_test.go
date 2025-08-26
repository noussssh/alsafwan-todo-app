package services

import (
	"testing"
	"time"

	"alsafwanmarine.com/todo-app/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	
	err = db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.UserActivity{},
		&models.PasswordResetEvent{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	
	return db
}

func TestAuthServiceLogin(t *testing.T) {
	db := setupTestDB(t)
	
	sessionService := NewSessionService(db)
	activityService := NewActivityService(db)
	authService := NewAuthService(db, sessionService, activityService)
	
	user := &models.User{
		Email:   "test@example.com",
		Name:    "Test User",
		Role:    models.RoleSalesperson,
		Enabled: true,
	}
	user.SetPassword("testpassword123")
	
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	credentials := LoginCredentials{
		Email:    "test@example.com",
		Password: "testpassword123",
	}
	
	result, err := authService.Login(credentials, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	
	if result.User.ID != user.ID {
		t.Error("Returned user ID doesn't match")
	}
	
	if result.Token == "" {
		t.Error("Token should not be empty")
	}
	
	if result.Session == nil {
		t.Error("Session should not be nil")
	}
}

func TestAuthServiceLoginInvalidCredentials(t *testing.T) {
	db := setupTestDB(t)
	
	sessionService := NewSessionService(db)
	activityService := NewActivityService(db)
	authService := NewAuthService(db, sessionService, activityService)
	
	credentials := LoginCredentials{
		Email:    "nonexistent@example.com",
		Password: "wrongpassword",
	}
	
	_, err := authService.Login(credentials, "127.0.0.1", "test-agent")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthServiceLoginDisabledUser(t *testing.T) {
	db := setupTestDB(t)
	
	sessionService := NewSessionService(db)
	activityService := NewActivityService(db)
	authService := NewAuthService(db, sessionService, activityService)
	
	user := &models.User{
		Email:   "disabled@example.com",
		Name:    "Disabled User",
		Role:    models.RoleSalesperson,
		Enabled: true, // Create as enabled first
	}
	user.SetPassword("testpassword123")
	
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Now disable the user
	user.Enabled = false
	if err := db.Save(user).Error; err != nil {
		t.Fatalf("Failed to disable test user: %v", err)
	}
	
	credentials := LoginCredentials{
		Email:    "disabled@example.com",
		Password: "testpassword123",
	}
	
	result, err := authService.Login(credentials, "127.0.0.1", "test-agent")
	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials for disabled user, got %v", err)
		if result != nil {
			t.Errorf("Result should be nil, got user: %+v", result.User)
		}
	}
}

func TestAuthServiceGetCurrentUser(t *testing.T) {
	db := setupTestDB(t)
	
	sessionService := NewSessionService(db)
	activityService := NewActivityService(db)
	authService := NewAuthService(db, sessionService, activityService)
	
	user := &models.User{
		Email:   "test@example.com",
		Name:    "Test User",
		Role:    models.RoleSalesperson,
		Enabled: true,
	}
	
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	session, token, err := sessionService.CreateSession(user, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	
	currentUser, err := authService.GetCurrentUser(token)
	if err != nil {
		t.Fatalf("GetCurrentUser failed: %v", err)
	}
	
	if currentUser.ID != user.ID {
		t.Error("Current user ID doesn't match")
	}
	
	var updatedSession models.Session
	db.First(&updatedSession, session.ID)
	
	if updatedSession.ExpiresAt.Before(time.Now().Add(25 * time.Minute)) {
		t.Error("Session should have been extended")
	}
}

func TestAuthServiceLogout(t *testing.T) {
	db := setupTestDB(t)
	
	sessionService := NewSessionService(db)
	activityService := NewActivityService(db)
	authService := NewAuthService(db, sessionService, activityService)
	
	user := &models.User{
		Email:   "test@example.com",
		Name:    "Test User",
		Role:    models.RoleSalesperson,
		Enabled: true,
	}
	
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	_, token, err := sessionService.CreateSession(user, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	
	err = authService.Logout(token, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
	
	_, err = authService.GetCurrentUser(token)
	if err == nil {
		t.Error("Should not be able to get current user after logout")
	}
}

func TestAuthServiceChangePassword(t *testing.T) {
	db := setupTestDB(t)
	
	sessionService := NewSessionService(db)
	activityService := NewActivityService(db)
	authService := NewAuthService(db, sessionService, activityService)
	
	user := &models.User{
		Email:   "test@example.com",
		Name:    "Test User",
		Role:    models.RoleSalesperson,
		Enabled: true,
	}
	user.SetPassword("oldpassword123")
	
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	err := authService.ChangePassword(user.ID, "oldpassword123", "newpassword123")
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}
	
	var updatedUser models.User
	db.First(&updatedUser, user.ID)
	
	if !updatedUser.CheckPassword("newpassword123") {
		t.Error("New password should be valid")
	}
	
	if updatedUser.CheckPassword("oldpassword123") {
		t.Error("Old password should no longer be valid")
	}
}