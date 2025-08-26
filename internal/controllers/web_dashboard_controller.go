package controllers

import (
	"time"

	"alsafwanmarine.com/todo-app/internal/middleware"
	"alsafwanmarine.com/todo-app/internal/models"
	"alsafwanmarine.com/todo-app/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WebDashboardController struct {
	db              *gorm.DB
	activityService *services.ActivityService
}

func NewWebDashboardController(db *gorm.DB, activityService *services.ActivityService) *WebDashboardController {
	return &WebDashboardController{
		db:              db,
		activityService: activityService,
	}
}

type DashboardStats struct {
	TotalUsers         int64
	ActiveUsers        int64
	SessionsToday      int64
	FailedLoginsToday  int64
}

func (dc *WebDashboardController) ShowDashboard(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}

	// Calculate statistics
	stats := DashboardStats{}
	
	// Total users
	dc.db.Model(&models.User{}).Count(&stats.TotalUsers)
	
	// Active users
	dc.db.Model(&models.User{}).Where("enabled = ?", true).Count(&stats.ActiveUsers)
	
	// Today's sessions (approximate based on login activities)
	today := time.Now().Truncate(24 * time.Hour)
	dc.db.Model(&models.UserActivity{}).
		Where("activity_type = ? AND performed_at >= ?", "login", today).
		Count(&stats.SessionsToday)
	
	// Today's failed logins
	dc.db.Model(&models.UserActivity{}).
		Where("activity_type = ? AND performed_at >= ?", "failed_login", today).
		Count(&stats.FailedLoginsToday)

	// Get recent activities
	recentActivities, _ := dc.activityService.GetAllActivities(10)

	c.HTML(200, "base.html", gin.H{
		"Title":       "Dashboard",
		"User":        user,
		"ActiveNav":   "dashboard",
		"Stats":       stats,
		"RecentActivities": recentActivities,
	})
}