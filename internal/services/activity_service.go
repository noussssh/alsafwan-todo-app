package services

import (
	"database/sql"
	"encoding/json"
	"time"

	"alsafwanmarine.com/todo-app/internal/models"
	"gorm.io/gorm"
)

type ActivityService struct {
	db *gorm.DB
}

func NewActivityService(db *gorm.DB) *ActivityService {
	return &ActivityService{db: db}
}

func (s *ActivityService) LogActivity(userID *uint, activityType, ipAddress, userAgent string, metadata map[string]interface{}) error {
	var metadataJSON sql.NullString
	if metadata != nil {
		bytes, err := json.Marshal(metadata)
		if err == nil {
			metadataJSON = sql.NullString{String: string(bytes), Valid: true}
		}
	}
	
	activity := &models.UserActivity{
		UserID:      userID,
		ActivityType: activityType,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Metadata:    metadataJSON,
		PerformedAt: time.Now(),
	}
	
	return s.db.Create(activity).Error
}

func (s *ActivityService) LogLogin(user *models.User, ipAddress, userAgent string) error {
	metadata := map[string]interface{}{
		"user_id":   user.ID,
		"user_name": user.Name,
		"user_role": user.Role.String(),
	}
	return s.LogActivity(&user.ID, "login", ipAddress, userAgent, metadata)
}

func (s *ActivityService) LogLogout(user *models.User, ipAddress, userAgent string) error {
	metadata := map[string]interface{}{
		"user_id":   user.ID,
		"user_name": user.Name,
	}
	return s.LogActivity(&user.ID, "logout", ipAddress, userAgent, metadata)
}

func (s *ActivityService) LogFailedLogin(userID *uint, email, ipAddress, userAgent string) error {
	metadata := map[string]interface{}{
		"attempted_email": email,
	}
	if userID != nil {
		metadata["user_id"] = *userID
	}
	return s.LogActivity(userID, "failed_login", ipAddress, userAgent, metadata)
}

func (s *ActivityService) LogPasswordChange(user *models.User, ipAddress, userAgent string) error {
	metadata := map[string]interface{}{
		"user_id":   user.ID,
		"user_name": user.Name,
	}
	return s.LogActivity(&user.ID, "password_change", ipAddress, userAgent, metadata)
}

func (s *ActivityService) LogPageView(user *models.User, page, ipAddress, userAgent string) error {
	metadata := map[string]interface{}{
		"page":      page,
		"user_id":   user.ID,
		"user_name": user.Name,
	}
	return s.LogActivity(&user.ID, "page_view", ipAddress, userAgent, metadata)
}

func (s *ActivityService) LogUserCRUD(performingUser *models.User, targetUser *models.User, action, ipAddress, userAgent string) error {
	metadata := map[string]interface{}{
		"performing_user_id":   performingUser.ID,
		"performing_user_name": performingUser.Name,
		"target_user_id":       targetUser.ID,
		"target_user_name":     targetUser.Name,
		"action":              action,
	}
	return s.LogActivity(&performingUser.ID, "user_crud", ipAddress, userAgent, metadata)
}

func (s *ActivityService) GetUserActivities(userID uint, limit int) ([]models.UserActivity, error) {
	var activities []models.UserActivity
	query := s.db.Where("user_id = ?", userID).Order("performed_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&activities).Error
	return activities, err
}

func (s *ActivityService) GetAllActivities(limit int) ([]models.UserActivity, error) {
	var activities []models.UserActivity
	query := s.db.Preload("User").Order("performed_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&activities).Error
	return activities, err
}