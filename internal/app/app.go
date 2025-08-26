package app

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"alsafwanmarine.com/todo-app/internal/config"
	"alsafwanmarine.com/todo-app/internal/controllers"
	"alsafwanmarine.com/todo-app/internal/middleware"
	"alsafwanmarine.com/todo-app/internal/models"
	"alsafwanmarine.com/todo-app/internal/services"
	"github.com/gin-gonic/gin"
)

type Application struct {
	Database             *config.Database
	AuthService          *services.AuthService
	SessionService       *services.SessionService
	ActivityService      *services.ActivityService
	PasswordResetService *services.PasswordResetService
	
	WebAuthController      *controllers.WebAuthController
	WebDashboardController *controllers.WebDashboardController
	WebUserController      *controllers.WebUserController
	
	AuthMiddleware *middleware.AuthMiddleware
	WebMiddleware  *middleware.WebMiddleware
}

func New(dbPath string) (*Application, error) {
	database, err := config.NewDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	
	sessionService := services.NewSessionService(database.DB)
	activityService := services.NewActivityService(database.DB)
	passwordResetService := services.NewPasswordResetService(database.DB, activityService)
	authService := services.NewAuthService(database.DB, sessionService, activityService)
	
	webAuthController := controllers.NewWebAuthController(authService)
	webDashboardController := controllers.NewWebDashboardController(database.DB, activityService)
	webUserController := controllers.NewWebUserController(database.DB, activityService, passwordResetService)
	
	authMiddleware := middleware.NewAuthMiddleware(authService, activityService)
	webMiddleware := middleware.NewWebMiddleware()
	
	app := &Application{
		Database:                database,
		AuthService:             authService,
		SessionService:          sessionService,
		ActivityService:         activityService,
		PasswordResetService:    passwordResetService,
		WebAuthController:       webAuthController,
		WebDashboardController:  webDashboardController,
		WebUserController:       webUserController,
		AuthMiddleware:          authMiddleware,
		WebMiddleware:           webMiddleware,
	}
	
	if err := database.Seed(); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	}
	
	go app.startBackgroundTasks()
	
	return app, nil
}

func (app *Application) SetupRoutes(r *gin.Engine) {
	// Try to load HTML templates, but don't fail if they don't exist
	templatePattern := "templates/**/*.html"
	if _, err := os.Stat("./templates"); err == nil {
		// Templates directory exists, try to load templates
		// Use a more comprehensive search to check if any templates exist
		matches, err := filepath.Glob("templates/*/*.html")
		if err == nil && len(matches) > 0 {
			r.LoadHTMLGlob(templatePattern)
			log.Printf("Loaded template files from templates directory")
		} else {
			log.Printf("Templates directory exists but no HTML files found")
		}
	} else {
		log.Printf("Templates directory not found: %v", err)
	}
	
	// Try to serve static files, but don't fail if directory doesn't exist
	if _, err := os.Stat("./static"); err == nil {
		r.Static("/static", "./static")
	} else {
		log.Printf("Static directory not found: %v", err)
	}
	
	if _, err := os.Stat("./static/favicon.ico"); err == nil {
		r.StaticFile("/favicon.ico", "./static/favicon.ico")
	}

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.InputSanitizer())
	r.Use(app.WebMiddleware.FlashMessages())
	r.Use(app.AuthMiddleware.OptionalAuth())
	r.Use(app.AuthMiddleware.ActivityLogger())

	// Authentication routes
	r.GET("/login", app.WebAuthController.ShowLogin)
	r.POST("/login", middleware.LoginRateLimit(), app.WebAuthController.HandleLogin)
	r.GET("/logout", app.WebAuthController.HandleLogout)

	// Protected routes
	protected := r.Group("/")
	protected.Use(middleware.RequireWebAuth())
	{
		// Dashboard
		protected.GET("/", middleware.SetActiveNav("dashboard"), app.WebDashboardController.ShowDashboard)
		
		// Profile routes
		protected.GET("/profile", middleware.SetActiveNav("profile"), app.WebAuthController.ShowProfile)
		protected.GET("/profile/password", middleware.SetActiveNav("profile"), app.WebAuthController.ShowChangePassword)
		protected.POST("/profile/password", app.WebAuthController.HandleChangePassword)

		// User management routes
		userRoutes := protected.Group("/users")
		userRoutes.Use(middleware.RequireWebRole(models.RoleManager))
		userRoutes.Use(middleware.SetActiveNav("users"))
		{
			userRoutes.GET("/", app.WebUserController.ListUsers)
			userRoutes.GET("/:id", app.WebUserController.ShowUser)
			userRoutes.GET("/new", app.WebUserController.ShowCreateUser)
			userRoutes.POST("/", app.WebUserController.HandleCreateUser)
			userRoutes.GET("/:id/edit", app.WebUserController.ShowEditUser)
			userRoutes.POST("/:id", app.WebUserController.HandleEditUser)
			userRoutes.GET("/:id/delete", app.WebUserController.HandleDeleteUser)
			userRoutes.GET("/:id/toggle-status", app.WebUserController.HandleToggleStatus)
			userRoutes.POST("/:id/reset-password", app.WebUserController.HandleResetPassword)
		}
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}

func (app *Application) startBackgroundTasks() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		if err := app.SessionService.CleanupExpiredSessions(); err != nil {
			log.Printf("Failed to cleanup expired sessions: %v", err)
		}
		
		if err := app.PasswordResetService.AutoResetExpiredPasswords(); err != nil {
			log.Printf("Failed to auto-reset expired passwords: %v", err)
		}
		
		if err := app.PasswordResetService.AutoResetInactiveUsers(); err != nil {
			log.Printf("Failed to auto-reset inactive users: %v", err)
		}
	}
}

func (app *Application) Close() error {
	return app.Database.Close()
}