package controllers

import (
	"net/http"
	"strings"

	"alsafwanmarine.com/todo-app/internal/middleware"
	"alsafwanmarine.com/todo-app/internal/models"
	"alsafwanmarine.com/todo-app/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PasswordResetController struct {
	db                   *gorm.DB
	passwordResetService *services.PasswordResetService
}

func NewPasswordResetController(db *gorm.DB, passwordResetService *services.PasswordResetService) *PasswordResetController {
	return &PasswordResetController{
		db:                   db,
		passwordResetService: passwordResetService,
	}
}

type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (prc *PasswordResetController) RequestPasswordReset(c *gin.Context) {
	var req RequestPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	normalizeEmail := func(email string) string {
		return strings.ToLower(strings.TrimSpace(email))
	}
	
	var user models.User
	if err := prc.db.Where("email = ?", normalizeEmail(req.Email)).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent"})
		return
	}
	
	if !user.Enabled {
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent"})
		return
	}
	
	resetEvent, err := prc.passwordResetService.CreateResetEvent(
		user.ID,
		nil,
		"User requested password reset",
		models.ResetTypeManual,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reset request"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset link has been sent",
		"token": *resetEvent.Token,
	})
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (prc *PasswordResetController) ResetPasswordWithToken(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := prc.passwordResetService.ResetPasswordWithToken(req.Token, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (prc *PasswordResetController) GetPasswordResetEvents(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	if currentUser.Role != models.RoleAdmin && currentUser.Role != models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	
	events, err := prc.passwordResetService.GetAllResetEvents(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reset events"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"events": events})
}