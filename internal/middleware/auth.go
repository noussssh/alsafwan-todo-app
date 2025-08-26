package middleware

import (
	"net/http"
	"strings"

	"alsafwanmarine.com/todo-app/internal/models"
	"alsafwanmarine.com/todo-app/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService *services.AuthService
	activityService *services.ActivityService
}

func NewAuthMiddleware(authService *services.AuthService, activityService *services.ActivityService) *AuthMiddleware {
	return &AuthMiddleware{
		authService:     authService,
		activityService: activityService,
	}
}

func (m *AuthMiddleware) getSessionToken(c *gin.Context) string {
	if token := c.GetHeader("Authorization"); token != "" {
		if strings.HasPrefix(token, "Bearer ") {
			return strings.TrimPrefix(token, "Bearer ")
		}
	}
	
	if token, err := c.Cookie("session_token"); err == nil {
		return token
	}
	
	return ""
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.getSessionToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		
		user, err := m.authService.RequireAuth(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			c.Abort()
			return
		}
		
		c.Set("current_user", user)
		c.Set("session_token", token)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.getSessionToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		
		user, err := m.authService.RequireRole(token, models.RoleAdmin)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		
		c.Set("current_user", user)
		c.Set("session_token", token)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireManagerOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.getSessionToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		
		user, err := m.authService.RequireRoleOrHigher(token, models.RoleManager)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Manager or Admin access required"})
			c.Abort()
			return
		}
		
		c.Set("current_user", user)
		c.Set("session_token", token)
		c.Next()
	}
}

func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.getSessionToken(c)
		if token != "" {
			if user, err := m.authService.GetCurrentUser(token); err == nil {
				c.Set("current_user", user)
				c.Set("session_token", token)
			}
		}
		c.Next()
	}
}

func (m *AuthMiddleware) ActivityLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		if c.Request.Method == "GET" && !strings.Contains(c.GetHeader("Accept"), "application/json") {
			if user, exists := c.Get("current_user"); exists {
				if u, ok := user.(*models.User); ok {
					m.activityService.LogPageView(u, c.Request.URL.Path, c.ClientIP(), c.Request.UserAgent())
				}
			}
		}
	}
}

func GetCurrentUser(c *gin.Context) *models.User {
	if user, exists := c.Get("current_user"); exists {
		if u, ok := user.(*models.User); ok {
			return u
		}
	}
	return nil
}

func GetSessionToken(c *gin.Context) string {
	if token, exists := c.Get("session_token"); exists {
		if t, ok := token.(string); ok {
			return t
		}
	}
	return ""
}