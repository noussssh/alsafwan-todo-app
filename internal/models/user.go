package models

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRole int

const (
	RoleAdmin UserRole = iota
	RoleManager
	RoleSalesperson
)

func (r UserRole) String() string {
	switch r {
	case RoleAdmin:
		return "admin"
	case RoleManager:
		return "manager"
	case RoleSalesperson:
		return "salesperson"
	default:
		return "unknown"
	}
}

type User struct {
	ID                     uint           `gorm:"primaryKey" json:"id"`
	Email                  string         `gorm:"uniqueIndex;not null" json:"email"`
	Name                   string         `gorm:"not null;size:100" json:"name"`
	PasswordDigest         string         `gorm:"not null" json:"-"`
	Role                   UserRole       `gorm:"not null;default:2" json:"role"`
	Company                *string        `gorm:"size:100" json:"company"`
	Enabled                bool           `gorm:"default:true" json:"enabled"`
	LastSignInAt           *time.Time     `json:"last_sign_in_at"`
	CurrentSignInAt        *time.Time     `json:"current_sign_in_at"`
	SignInCount            int            `gorm:"default:0" json:"sign_in_count"`
	PasswordResetAt        *time.Time     `json:"password_reset_at"`
	PasswordExpiresAt      *time.Time     `json:"password_expires_at"`
	ManagedCustomersCount  int            `gorm:"default:0" json:"managed_customers_count"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	
	Sessions               []Session      `gorm:"foreignKey:UserID"`
	Activities             []UserActivity `gorm:"foreignKey:UserID"`
	PasswordResetEvents    []PasswordResetEvent `gorm:"foreignKey:UserID"`
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordDigest = string(hashedPassword)
	now := time.Now()
	u.PasswordResetAt = &now
	expiry := now.Add(30 * 24 * time.Hour)
	u.PasswordExpiresAt = &expiry
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordDigest), []byte(password))
	return err == nil
}

func (u *User) IsPasswordExpired() bool {
	if u.PasswordExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.PasswordExpiresAt)
}

func (u *User) ShouldResetForInactivity() bool {
	if u.LastSignInAt == nil {
		return false
	}
	tenDaysAgo := time.Now().Add(-10 * 24 * time.Hour)
	return u.LastSignInAt.Before(tenDaysAgo)
}

func (u *User) UpdateSignInInfo() {
	now := time.Now()
	u.LastSignInAt = u.CurrentSignInAt
	u.CurrentSignInAt = &now
	u.SignInCount++
}

func (u *User) CanManageUser(targetUser *User) bool {
	switch u.Role {
	case RoleAdmin:
		return true
	case RoleManager:
		return targetUser.Role == RoleSalesperson
	default:
		return false
	}
}

func (u *User) CanDisableUser(targetUser *User) bool {
	if u.ID == targetUser.ID {
		return false
	}
	
	if targetUser.Role != RoleSalesperson {
		return false
	}
	
	return u.CanManageUser(targetUser)
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.Email = normalizeEmail(u.Email)
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.Email = normalizeEmail(u.Email)
	return nil
}

type Session struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"-"`
	IPAddress string    `gorm:"size:45" json:"ip_address"`
	UserAgent string    `gorm:"size:500" json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	User      User      `gorm:"foreignKey:UserID"`
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) Extend() {
	s.ExpiresAt = time.Now().Add(30 * time.Minute)
}

type UserActivity struct {
	ID              uint            `gorm:"primaryKey" json:"id"`
	UserID          *uint           `gorm:"index" json:"user_id"`
	ActivityType    string          `gorm:"not null;size:50" json:"activity_type"`
	SubjectType     *string         `gorm:"size:50" json:"subject_type"`
	SubjectID       *uint           `json:"subject_id"`
	IPAddress       string          `gorm:"size:45" json:"ip_address"`
	UserAgent       string          `gorm:"size:500" json:"user_agent"`
	SessionDuration *int            `json:"session_duration"`
	Metadata        sql.NullString  `gorm:"type:json" json:"metadata"`
	PerformedAt     time.Time       `json:"performed_at"`
	
	User            *User           `gorm:"foreignKey:UserID"`
}

type ResetType string

const (
	ResetTypeManual             ResetType = "manual"
	ResetTypeAutomaticExpiry    ResetType = "automatic_expiry"
	ResetTypeAutomaticInactivity ResetType = "automatic_inactivity"
)

type PasswordResetEvent struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	AdminID    *uint     `gorm:"index" json:"admin_id"`
	Reason     string    `gorm:"size:500" json:"reason"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	UserAgent  string    `gorm:"size:500" json:"user_agent"`
	Success    bool      `gorm:"default:false" json:"success"`
	ResetType  ResetType `gorm:"not null" json:"reset_type"`
	Token      *string   `gorm:"uniqueIndex;size:100" json:"-"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	
	User       User      `gorm:"foreignKey:UserID"`
	Admin      *User     `gorm:"foreignKey:AdminID"`
}

func (p *PasswordResetEvent) IsExpired() bool {
	if p.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*p.ExpiresAt)
}