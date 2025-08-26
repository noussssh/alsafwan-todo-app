package services

import (
	"errors"
	"strings"

	"alsafwanmarine.com/todo-app/internal/models"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
	sessionService *SessionService
	activityService *ActivityService
}

func NewAuthService(db *gorm.DB, sessionService *SessionService, activityService *ActivityService) *AuthService {
	return &AuthService{
		db:             db,
		sessionService: sessionService,
		activityService: activityService,
	}
}

type LoginCredentials struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResult struct {
	User    *models.User    `json:"user"`
	Session *models.Session `json:"session"`
	Token   string         `json:"token"`
}

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserDisabled      = errors.New("user account is disabled")
	ErrRateLimited       = errors.New("too many login attempts")
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *AuthService) Login(credentials LoginCredentials, ipAddress, userAgent string) (*LoginResult, error) {
	var user models.User
	if err := s.db.Where("email = ?", normalizeEmail(credentials.Email)).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.activityService.LogFailedLogin(nil, credentials.Email, ipAddress, userAgent)
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	
	if !user.Enabled {
		s.activityService.LogFailedLogin(&user.ID, credentials.Email, ipAddress, userAgent)
		return nil, ErrInvalidCredentials
	}
	
	if !user.CheckPassword(credentials.Password) {
		s.activityService.LogFailedLogin(&user.ID, credentials.Email, ipAddress, userAgent)
		return nil, ErrInvalidCredentials
	}
	
	if user.IsPasswordExpired() {
		return nil, errors.New("password has expired")
	}
	
	user.UpdateSignInInfo()
	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	
	session, token, err := s.sessionService.CreateSession(&user, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}
	
	s.activityService.LogLogin(&user, ipAddress, userAgent)
	
	return &LoginResult{
		User:    &user,
		Session: session,
		Token:   token,
	}, nil
}

func (s *AuthService) Logout(sessionToken string, ipAddress, userAgent string) error {
	session, err := s.sessionService.GetSessionByToken(sessionToken)
	if err != nil {
		return err
	}
	
	if session == nil {
		return errors.New("invalid session")
	}
	
	var user models.User
	if err := s.db.First(&user, session.UserID).Error; err != nil {
		return err
	}
	
	if err := s.sessionService.DestroySession(sessionToken); err != nil {
		return err
	}
	
	s.activityService.LogLogout(&user, ipAddress, userAgent)
	
	return nil
}

func (s *AuthService) GetCurrentUser(sessionToken string) (*models.User, error) {
	session, err := s.sessionService.GetSessionByToken(sessionToken)
	if err != nil {
		return nil, err
	}
	
	if session == nil || session.IsExpired() {
		return nil, errors.New("invalid or expired session")
	}
	
	var user models.User
	if err := s.db.First(&user, session.UserID).Error; err != nil {
		return nil, err
	}
	
	if !user.Enabled {
		s.sessionService.DestroySession(sessionToken)
		return nil, errors.New("user account is disabled")
	}
	
	session.Extend()
	s.db.Save(session)
	
	return &user, nil
}

func (s *AuthService) IsAuthenticated(sessionToken string) bool {
	user, err := s.GetCurrentUser(sessionToken)
	return err == nil && user != nil
}

func (s *AuthService) RequireAuth(sessionToken string) (*models.User, error) {
	return s.GetCurrentUser(sessionToken)
}

func (s *AuthService) RequireRole(sessionToken string, requiredRole models.UserRole) (*models.User, error) {
	user, err := s.RequireAuth(sessionToken)
	if err != nil {
		return nil, err
	}
	
	if user.Role != requiredRole {
		return nil, errors.New("insufficient permissions")
	}
	
	return user, nil
}

func (s *AuthService) RequireRoleOrHigher(sessionToken string, minRole models.UserRole) (*models.User, error) {
	user, err := s.RequireAuth(sessionToken)
	if err != nil {
		return nil, err
	}
	
	if int(user.Role) > int(minRole) {
		return nil, errors.New("insufficient permissions")
	}
	
	return user, nil
}

func (s *AuthService) ChangePassword(userID uint, currentPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return err
	}
	
	if !user.CheckPassword(currentPassword) {
		return errors.New("current password is incorrect")
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
	
	s.activityService.LogPasswordChange(&user, "", "")
	
	return nil
}