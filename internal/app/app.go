package app

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"time"

	"alsafwanmarine.com/todo-app/internal/cache"
	"alsafwanmarine.com/todo-app/internal/config"
	"alsafwanmarine.com/todo-app/internal/controllers"
	"alsafwanmarine.com/todo-app/internal/middleware"
	"alsafwanmarine.com/todo-app/internal/models"
	"alsafwanmarine.com/todo-app/internal/services"
	"github.com/gin-gonic/gin"
)

type Application struct {
	Database             *config.Database
	Cache               *cache.Cache
	AuthService          *services.AuthService
	SessionService       *services.SessionService
	ActivityService      *services.ActivityService
	PasswordResetService *services.PasswordResetService
	CachedStatsService   *services.CachedStatsService
	
	WebAuthController      *controllers.WebAuthController
	WebDashboardController *controllers.WebDashboardController
	WebUserController      *controllers.WebUserController
	
	AuthMiddleware *middleware.AuthMiddleware
	WebMiddleware  *middleware.WebMiddleware
	
	templatesFS embed.FS
	staticFS    embed.FS
}

func New(dbPath string, templatesFS, staticFS embed.FS) (*Application, error) {
	database, err := config.NewDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	
	// Initialize cache with 5-minute cleanup interval
	appCache := cache.New(5 * time.Minute)
	
	sessionService := services.NewSessionService(database.DB)
	activityService := services.NewActivityService(database.DB)
	passwordResetService := services.NewPasswordResetService(database.DB, activityService)
	authService := services.NewAuthService(database.DB, sessionService, activityService)
	cachedStatsService := services.NewCachedStatsService(database.DB, appCache)
	
	webAuthController := controllers.NewWebAuthController(authService)
	webDashboardController := controllers.NewWebDashboardController(database.DB, activityService)
	webUserController := controllers.NewWebUserController(database.DB, activityService, passwordResetService)
	
	authMiddleware := middleware.NewAuthMiddleware(authService, activityService)
	webMiddleware := middleware.NewWebMiddleware()
	
	app := &Application{
		Database:                database,
		Cache:                   appCache,
		AuthService:             authService,
		SessionService:          sessionService,
		ActivityService:         activityService,
		PasswordResetService:    passwordResetService,
		CachedStatsService:      cachedStatsService,
		WebAuthController:       webAuthController,
		WebDashboardController:  webDashboardController,
		WebUserController:       webUserController,
		AuthMiddleware:          authMiddleware,
		WebMiddleware:           webMiddleware,
		templatesFS:             templatesFS,
		staticFS:                staticFS,
	}
	
	if err := database.Seed(); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	}
	
	go app.startBackgroundTasks()
	
	return app, nil
}

func (app *Application) SetupRoutes(r *gin.Engine) {
	// Load embedded templates
	templ := template.Must(template.New("").ParseFS(app.templatesFS, "templates/**/*.html"))
	r.SetHTMLTemplate(templ)
	log.Printf("Loaded embedded template files")
	
	// Serve embedded static files
	staticSubFS, err := fs.Sub(app.staticFS, "static")
	if err != nil {
		log.Printf("Warning: Could not create static file system: %v", err)
	} else {
		r.StaticFS("/static", http.FS(staticSubFS))
		log.Printf("Serving embedded static files")
	}
	
	// Serve favicon from embedded files
	r.GET("/favicon.ico", func(c *gin.Context) {
		data, err := app.staticFS.ReadFile("static/favicon.ico")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Header("Content-Type", "image/x-icon")
		c.Data(http.StatusOK, "image/x-icon", data)
	})

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.PerformanceLogger())
	r.Use(middleware.RequestSizeLimit(10 << 20)) // 10MB limit
	r.Use(middleware.Gzip(middleware.DefaultCompression))
	r.Use(middleware.StaticFileHeaders())
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

	// Health check with performance metrics
	r.GET("/health", middleware.HealthCheck())
	
	// Performance metrics endpoint (could be restricted in production)
	r.GET("/metrics", middleware.HealthCheck())
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
	if app.Cache != nil {
		app.Cache.Close()
	}
	return app.Database.Close()
}