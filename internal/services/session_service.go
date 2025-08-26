package services

import (
	"time"

	"alsafwanmarine.com/todo-app/internal/models"
	"gorm.io/gorm"
)

type SessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

func (s *SessionService) CreateSession(user *models.User, ipAddress, userAgent string) (*models.Session, string, error) {
	token, err := models.GenerateSecureToken()
	if err != nil {
		return nil, "", err
	}
	
	session := &models.Session{
		UserID:    user.ID,
		Token:     token,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	
	if err := s.db.Create(session).Error; err != nil {
		return nil, "", err
	}
	
	return session, token, nil
}

func (s *SessionService) GetSessionByToken(token string) (*models.Session, error) {
	var session models.Session
	if err := s.db.Where("token = ?", token).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &session, nil
}

func (s *SessionService) DestroySession(token string) error {
	return s.db.Where("token = ?", token).Delete(&models.Session{}).Error
}

func (s *SessionService) DestroyUserSessions(userID uint) error {
	return s.db.Where("user_id = ?", userID).Delete(&models.Session{}).Error
}

func (s *SessionService) CleanupExpiredSessions() error {
	return s.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error
}

func (s *SessionService) ExtendSession(token string) error {
	return s.db.Model(&models.Session{}).
		Where("token = ?", token).
		Update("expires_at", time.Now().Add(30*time.Minute)).Error
}