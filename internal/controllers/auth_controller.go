package controllers

import (
	"net/http"
	"time"

	"alsafwanmarine.com/todo-app/internal/middleware"
	"alsafwanmarine.com/todo-app/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (ac *AuthController) Login(c *gin.Context) {
	var credentials services.LoginCredentials
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	result, err := ac.authService.Login(credentials, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		case services.ErrRateLimited:
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		}
		return
	}
	
	c.SetCookie(
		"session_token",
		result.Token,
		int(30*time.Minute.Seconds()),
		"/",
		"",
		true,
		true,
	)
	
	c.JSON(http.StatusOK, gin.H{
		"user": result.User,
		"message": "Login successful",
	})
}

func (ac *AuthController) Logout(c *gin.Context) {
	token := middleware.GetSessionToken(c)
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active session"})
		return
	}
	
	err := ac.authService.Logout(token, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed"})
		return
	}
	
	c.SetCookie(
		"session_token",
		"",
		-1,
		"/",
		"",
		true,
		true,
	)
	
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func (ac *AuthController) GetCurrentUser(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"user": user})
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

func (ac *AuthController) ChangePassword(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := ac.authService.ChangePassword(user.ID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}