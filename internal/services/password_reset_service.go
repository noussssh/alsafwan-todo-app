package services

import (
	"errors"
	"time"

	"alsafwanmarine.com/todo-app/internal/models"
	"gorm.io/gorm"
)

type PasswordResetService struct {
	db *gorm.DB
	activityService *ActivityService
}

func NewPasswordResetService(db *gorm.DB, activityService *ActivityService) *PasswordResetService {
	return &PasswordResetService{
		db:             db,
		activityService: activityService,
	}
}

func (s *PasswordResetService) CreateResetEvent(userID uint, adminID *uint, reason string, resetType models.ResetType, ipAddress, userAgent string) (*models.PasswordResetEvent, error) {
	token, err := models.GenerateSecureToken()
	if err != nil {
		return nil, err
	}
	
	expiresAt := time.Now().Add(24 * time.Hour)
	
	resetEvent := &models.PasswordResetEvent{
		UserID:    userID,
		AdminID:   adminID,
		Reason:    reason,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   false,
		ResetType: resetType,
		Token:     &token,
		ExpiresAt: &expiresAt,
	}
	
	if err := s.db.Create(resetEvent).Error; err != nil {
		return nil, err
	}
	
	return resetEvent, nil
}

func (s *PasswordResetService) ResetPasswordWithToken(token, newPassword string) error {
	var resetEvent models.PasswordResetEvent
	if err := s.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&resetEvent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("invalid or expired reset token")
		}
		return err
	}
	
	var user models.User
	if err := s.db.First(&user, resetEvent.UserID).Error; err != nil {
		return err
	}
	
	if err := models.ValidatePassword(newPassword); err != nil {
		return err
	}
	
	if err := user.SetPassword(newPassword); err != nil {
		return err
	}
	
	if err := s.db.Save(&user).Error; err != nil {
		return err
	}
	
	resetEvent.Success = true
	if err := s.db.Save(&resetEvent).Error; err != nil {
		return err
	}
	
	s.activityService.LogPasswordChange(&user, "", "")
	
	return nil
}

func (s *PasswordResetService) ManualReset(userID uint, adminID uint, reason, ipAddress, userAgent string) (string, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return "", err
	}
	
	newPassword, err := models.GenerateStrongPassword()
	if err != nil {
		return "", err
	}
	
	if err := user.SetPassword(newPassword); err != nil {
		return "", err
	}
	
	if err := s.db.Save(&user).Error; err != nil {
		return "", err
	}
	
	resetEvent := &models.PasswordResetEvent{
		UserID:    userID,
		AdminID:   &adminID,
		Reason:    reason,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   true,
		ResetType: models.ResetTypeManual,
	}
	
	if err := s.db.Create(resetEvent).Error; err != nil {
		return "", err
	}
	
	var admin models.User
	s.db.First(&admin, adminID)
	s.activityService.LogUserCRUD(&admin, &user, "password_reset", ipAddress, userAgent)
	
	return newPassword, nil
}

func (s *PasswordResetService) AutoResetExpiredPasswords() error {
	var users []models.User
	if err := s.db.Where("password_expires_at IS NOT NULL AND password_expires_at < ?", time.Now()).Find(&users).Error; err != nil {
		return err
	}
	
	for _, user := range users {
		newPassword, err := models.GenerateStrongPassword()
		if err != nil {
			continue
		}
		
		if err := user.SetPassword(newPassword); err != nil {
			continue
		}
		
		if err := s.db.Save(&user).Error; err != nil {
			continue
		}
		
		resetEvent := &models.PasswordResetEvent{
			UserID:    user.ID,
			Reason:    "Password expired automatically",
			Success:   true,
			ResetType: models.ResetTypeAutomaticExpiry,
		}
		s.db.Create(resetEvent)
	}
	
	return nil
}

func (s *PasswordResetService) AutoResetInactiveUsers() error {
	var users []models.User
	tenDaysAgo := time.Now().Add(-10 * 24 * time.Hour)
	
	if err := s.db.Where("last_sign_in_at IS NOT NULL AND last_sign_in_at < ?", tenDaysAgo).Find(&users).Error; err != nil {
		return err
	}
	
	for _, user := range users {
		newPassword, err := models.GenerateStrongPassword()
		if err != nil {
			continue
		}
		
		if err := user.SetPassword(newPassword); err != nil {
			continue
		}
		
		if err := s.db.Save(&user).Error; err != nil {
			continue
		}
		
		resetEvent := &models.PasswordResetEvent{
			UserID:    user.ID,
			Reason:    "User inactive for more than 10 days",
			Success:   true,
			ResetType: models.ResetTypeAutomaticInactivity,
		}
		s.db.Create(resetEvent)
	}
	
	return nil
}

func (s *PasswordResetService) BulkResetPasswords(userIDs []uint, adminID uint, reason, ipAddress, userAgent string) (map[uint]string, error) {
	results := make(map[uint]string)
	
	for _, userID := range userIDs {
		newPassword, err := s.ManualReset(userID, adminID, reason, ipAddress, userAgent)
		if err != nil {
			continue
		}
		results[userID] = newPassword
	}
	
	return results, nil
}

func (s *PasswordResetService) GetResetEvents(userID uint) ([]models.PasswordResetEvent, error) {
	var events []models.PasswordResetEvent
	err := s.db.Where("user_id = ?", userID).
		Preload("Admin").
		Order("created_at DESC").
		Find(&events).Error
	return events, err
}

func (s *PasswordResetService) GetAllResetEvents(limit int) ([]models.PasswordResetEvent, error) {
	var events []models.PasswordResetEvent
	query := s.db.Preload("User").Preload("Admin").Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&events).Error
	return events, err
}