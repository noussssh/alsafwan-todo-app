package services

import (
	"fmt"
	"time"

	"alsafwanmarine.com/todo-app/internal/cache"
	"alsafwanmarine.com/todo-app/internal/models"
	"gorm.io/gorm"
)

// DashboardStats represents cached dashboard statistics
type DashboardStats struct {
	TotalUsers        int64     `json:"total_users"`
	ActiveUsers       int64     `json:"active_users"`
	SessionsToday     int64     `json:"sessions_today"`
	FailedLoginsToday int64     `json:"failed_logins_today"`
	LastUpdated       time.Time `json:"last_updated"`
}

// CachedStatsService provides cached statistics for dashboard
type CachedStatsService struct {
	db    *gorm.DB
	cache *cache.Cache
}

// NewCachedStatsService creates a new cached stats service
func NewCachedStatsService(db *gorm.DB, cache *cache.Cache) *CachedStatsService {
	return &CachedStatsService{
		db:    db,
		cache: cache,
	}
}

// GetDashboardStats returns cached dashboard statistics
func (css *CachedStatsService) GetDashboardStats() (*DashboardStats, error) {
	cacheKey := "dashboard_stats"
	
	// Try to get from cache first
	var stats DashboardStats
	found, err := css.cache.GetJSON(cacheKey, &stats)
	if err == nil && found {
		return &stats, nil
	}
	
	// Cache miss or error, fetch from database
	stats, err = css.fetchStatsFromDB()
	if err != nil {
		return nil, err
	}
	
	// Cache the results for 5 minutes
	css.cache.SetJSON(cacheKey, stats, 5*time.Minute)
	
	return &stats, nil
}

// GetUserList returns paginated and cached user list
func (css *CachedStatsService) GetUserList(page int, limit int, filters map[string]interface{}) ([]models.User, int64, error) {
	// Create cache key based on pagination and filters
	cacheKey := fmt.Sprintf("user_list_page_%d_limit_%d", page, limit)
	for k, v := range filters {
		cacheKey += fmt.Sprintf("_%s_%v", k, v)
	}
	
	type CachedUserList struct {
		Users      []models.User `json:"users"`
		TotalCount int64         `json:"total_count"`
		CachedAt   time.Time     `json:"cached_at"`
	}
	
	// Try cache first
	var cached CachedUserList
	found, err := css.cache.GetJSON(cacheKey, &cached)
	if err == nil && found {
		return cached.Users, cached.TotalCount, nil
	}
	
	// Cache miss, fetch from database
	users, totalCount, err := css.fetchUsersFromDB(page, limit, filters)
	if err != nil {
		return nil, 0, err
	}
	
	// Cache for 2 minutes (shorter TTL for user data as it changes more frequently)
	cached = CachedUserList{
		Users:      users,
		TotalCount: totalCount,
		CachedAt:   time.Now(),
	}
	css.cache.SetJSON(cacheKey, cached, 2*time.Minute)
	
	return users, totalCount, nil
}

// InvalidateUserCache removes user-related cache entries
func (css *CachedStatsService) InvalidateUserCache() {
	// This is a simple approach - in production, you'd want more sophisticated cache invalidation
	css.cache.Clear()
}

// InvalidateStatsCache removes stats cache
func (css *CachedStatsService) InvalidateStatsCache() {
	css.cache.Delete("dashboard_stats")
}

// fetchStatsFromDB retrieves statistics directly from database
func (css *CachedStatsService) fetchStatsFromDB() (DashboardStats, error) {
	var stats DashboardStats
	today := time.Now().Truncate(24 * time.Hour)
	
	// Use optimized raw queries for better performance
	queries := []struct {
		query string
		dest  *int64
		args  []interface{}
	}{
		{"SELECT COUNT(*) FROM users", &stats.TotalUsers, nil},
		{"SELECT COUNT(*) FROM users WHERE enabled = ?", &stats.ActiveUsers, []interface{}{true}},
		{"SELECT COUNT(*) FROM user_activities WHERE activity_type = ? AND performed_at >= ?", &stats.SessionsToday, []interface{}{"login", today}},
		{"SELECT COUNT(*) FROM user_activities WHERE activity_type = ? AND performed_at >= ?", &stats.FailedLoginsToday, []interface{}{"failed_login", today}},
	}
	
	for _, q := range queries {
		var result int64
		if q.args != nil {
			err := css.db.Raw(q.query, q.args...).Scan(&result).Error
			if err != nil {
				return stats, err
			}
		} else {
			err := css.db.Raw(q.query).Scan(&result).Error
			if err != nil {
				return stats, err
			}
		}
		*q.dest = result
	}
	
	stats.LastUpdated = time.Now()
	return stats, nil
}

// fetchUsersFromDB retrieves users with pagination and filters
func (css *CachedStatsService) fetchUsersFromDB(page int, limit int, filters map[string]interface{}) ([]models.User, int64, error) {
	var users []models.User
	var totalCount int64
	
	offset := (page - 1) * limit
	
	// Build query with filters
	query := css.db.Model(&models.User{})
	countQuery := css.db.Model(&models.User{})
	
	// Apply filters
	for key, value := range filters {
		switch key {
		case "role":
			query = query.Where("role = ?", value)
			countQuery = countQuery.Where("role = ?", value)
		case "enabled":
			query = query.Where("enabled = ?", value)
			countQuery = countQuery.Where("enabled = ?", value)
		case "search":
			searchTerm := fmt.Sprintf("%%%s%%", value)
			query = query.Where("name LIKE ? OR email LIKE ?", searchTerm, searchTerm)
			countQuery = countQuery.Where("name LIKE ? OR email LIKE ?", searchTerm, searchTerm)
		}
	}
	
	// Get total count
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}
	
	// Get users with pagination and only necessary fields
	if err := query.
		Select("id, name, email, role, company, enabled, created_at, last_sign_in_at").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	
	return users, totalCount, nil
}