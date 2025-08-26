package controllers

import (
	"net/http"
	"time"

	"alsafwanmarine.com/todo-app/internal/middleware"
	"alsafwanmarine.com/todo-app/internal/services"
	"github.com/gin-gonic/gin"
)

type WebAuthController struct {
	authService *services.AuthService
}

func NewWebAuthController(authService *services.AuthService) *WebAuthController {
	return &WebAuthController{
		authService: authService,
	}
}

func (ac *WebAuthController) ShowLogin(c *gin.Context) {
	// Try to render template, fall back to simple HTML if template fails
	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title": "Login",
		"Errors": make(map[string]string),
		"FormData": gin.H{
			"Email": "",
			"Remember": false,
		},
	})
}

func (ac *WebAuthController) HandleLogin(c *gin.Context) {
	// If already logged in, redirect to dashboard
	if user := middleware.GetCurrentUser(c); user != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	var credentials services.LoginCredentials
	
	credentials.Email = c.PostForm("email")
	credentials.Password = c.PostForm("password")
	remember := c.PostForm("remember") == "on"

	errors := make(map[string]string)
	formData := gin.H{
		"Email": credentials.Email,
		"Remember": remember,
	}

	// Basic validation
	if credentials.Email == "" {
		errors["Email"] = "Email is required"
	}
	if credentials.Password == "" {
		errors["Password"] = "Password is required"
	}

	if len(errors) > 0 {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"Title":    "Login",
			"Errors":   errors,
			"FormData": formData,
		})
		return
	}

	result, err := ac.authService.Login(credentials, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			errors["General"] = "Invalid email or password"
		case services.ErrRateLimited:
			errors["General"] = "Too many login attempts. Please try again later."
		default:
			errors["General"] = "Login failed. Please try again."
		}

		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"Title":    "Login",
			"Errors":   errors,
			"FormData": formData,
		})
		return
	}

	// Set session cookie
	cookieMaxAge := int(30 * time.Minute.Seconds())
	if remember {
		cookieMaxAge = int(24 * 7 * time.Hour.Seconds()) // 1 week
	}

	c.SetCookie(
		"session_token",
		result.Token,
		cookieMaxAge,
		"/",
		"",
		true,  // Secure
		true,  // HttpOnly
	)

	middleware.SetFlashSuccess(c, "Welcome back, "+result.User.Name+"!")
	c.Redirect(http.StatusFound, "/")
}

func (ac *WebAuthController) HandleLogout(c *gin.Context) {
	token := middleware.GetSessionToken(c)
	if token != "" {
		ac.authService.Logout(token, c.ClientIP(), c.Request.UserAgent())
	}

	// Clear session cookie
	c.SetCookie(
		"session_token",
		"",
		-1,
		"/",
		"",
		true,
		true,
	)

	middleware.SetFlashInfo(c, "You have been logged out successfully.")
	c.Redirect(http.StatusFound, "/login")
}

func (ac *WebAuthController) ShowProfile(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "base.html", gin.H{
		"Title":    "My Profile",
		"User":     user,
		"ActiveNav": "profile",
		"ViewUser": user,
	})
}

func (ac *WebAuthController) ShowChangePassword(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "base.html", gin.H{
		"Title":    "Change Password",
		"User":     user,
		"ActiveNav": "profile",
		"Errors":   make(map[string]string),
	})
}

func (ac *WebAuthController) HandleChangePassword(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	errors := make(map[string]string)

	// Validation
	if currentPassword == "" {
		errors["CurrentPassword"] = "Current password is required"
	}
	if newPassword == "" {
		errors["NewPassword"] = "New password is required"
	} else if len(newPassword) < 6 {
		errors["NewPassword"] = "Password must be at least 6 characters"
	}
	if newPassword != confirmPassword {
		errors["ConfirmPassword"] = "Passwords do not match"
	}

	if len(errors) > 0 {
		c.HTML(http.StatusBadRequest, "base.html", gin.H{
			"Title":    "Change Password",
			"User":     user,
			"ActiveNav": "profile",
			"Errors":   errors,
		})
		return
	}

	err := ac.authService.ChangePassword(user.ID, currentPassword, newPassword)
	if err != nil {
		errors["General"] = err.Error()
		c.HTML(http.StatusBadRequest, "base.html", gin.H{
			"Title":    "Change Password",
			"User":     user,
			"ActiveNav": "profile",
			"Errors":   errors,
		})
		return
	}

	middleware.SetFlashSuccess(c, "Password changed successfully!")
	c.Redirect(http.StatusFound, "/profile")
}