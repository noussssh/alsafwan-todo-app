package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CSRFConfig struct {
	TokenHeader string
	CookieName  string
	TokenLength int
	MaxAge      int
}

func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenHeader: "X-CSRF-Token",
		CookieName:  "csrf_token",
		TokenLength: 32,
		MaxAge:      3600,
	}
}

func generateCSRFToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func CSRFProtection(config ...CSRFConfig) gin.HandlerFunc {
	cfg := DefaultCSRFConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			token, err := generateCSRFToken(cfg.TokenLength)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSRF token"})
				c.Abort()
				return
			}
			
			c.SetCookie(
				cfg.CookieName,
				token,
				cfg.MaxAge,
				"/",
				"",
				true,
				false,
			)
			
			c.Header(cfg.TokenHeader, token)
			c.Next()
			return
		}
		
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" || c.Request.Method == "DELETE" {
			headerToken := c.GetHeader(cfg.TokenHeader)
			cookieToken, err := c.Cookie(cfg.CookieName)
			
			if err != nil || headerToken == "" || cookieToken == "" || headerToken != cookieToken {
				c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token validation failed"})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

func CSRFSkipper() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}