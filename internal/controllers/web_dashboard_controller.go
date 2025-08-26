package controllers

import (
	"time"

	"alsafwanmarine.com/todo-app/internal/middleware"
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

	// Calculate statistics with optimized queries
	stats := DashboardStats{}
	today := time.Now().Truncate(24 * time.Hour)
	
	// Use a single transaction to reduce database roundtrips
	tx := dc.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Execute optimized count queries using indexes
	queries := []struct {
		name  string
		query string
	}{
		{"total_users", "SELECT 'total_users' as query_type, COUNT(*) as count FROM users"},
		{"active_users", "SELECT 'active_users' as query_type, COUNT(*) as count FROM users WHERE enabled = 1"},
		{"sessions_today", "SELECT 'sessions_today' as query_type, COUNT(*) as count FROM user_activities WHERE activity_type = 'login' AND performed_at >= ?"},
		{"failed_logins_today", "SELECT 'failed_logins_today' as query_type, COUNT(*) as count FROM user_activities WHERE activity_type = 'failed_login' AND performed_at >= ?"},
	}
	
	for _, q := range queries {
		var result struct {
			QueryType string
			Count     int64
		}
		
		if q.name == "sessions_today" || q.name == "failed_logins_today" {
			tx.Raw(q.query, today).Scan(&result)
		} else {
			tx.Raw(q.query).Scan(&result)
		}
		
		switch result.QueryType {
		case "total_users":
			stats.TotalUsers = result.Count
		case "active_users":
			stats.ActiveUsers = result.Count
		case "sessions_today":
			stats.SessionsToday = result.Count
		case "failed_logins_today":
			stats.FailedLoginsToday = result.Count
		}
	}
	
	tx.Commit()

	// Get recent activities with preloading for better performance
	recentActivities, _ := dc.activityService.GetAllActivities(10)

	c.HTML(200, "base.html", gin.H{
		"Title":            "Dashboard",
		"User":             user,
		"ActiveNav":        "dashboard",
		"Stats":            stats,
		"RecentActivities": recentActivities,
	})
}