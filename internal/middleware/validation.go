package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	validate.RegisterValidation("strong_password", validateStrongPassword)
	validate.RegisterValidation("valid_company", validateValidCompany)
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 6 {
		return false
	}
	
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)
	
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func validateValidCompany(fl validator.FieldLevel) bool {
	company := fl.Field().String()
	if company == "" {
		return true
	}
	
	validCompanies := []string{
		"Al Safwan Marine",
		"Louis Safety",
		"Data Grid Labs",
	}
	
	for _, valid := range validCompanies {
		if company == valid {
			return true
		}
	}
	
	return false
}

func InputSanitizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("Content-Type") == "application/json" {
			c.Next()
			return
		}
		
		if err := c.Request.ParseForm(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
			c.Abort()
			return
		}
		
		for key, values := range c.Request.Form {
			for i, value := range values {
				c.Request.Form[key][i] = strings.TrimSpace(value)
			}
		}
		
		c.Next()
	}
}

func ValidateJSON(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(obj); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON data",
				"details": err.Error(),
			})
			c.Abort()
			return
		}
		
		if err := validate.Struct(obj); err != nil {
			var validationErrors []string
			for _, err := range err.(validator.ValidationErrors) {
				validationErrors = append(validationErrors, err.Error())
			}
			
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Validation failed",
				"details": validationErrors,
			})
			c.Abort()
			return
		}
		
		c.Set("validated_data", obj)
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		
		c.Next()
	}
}