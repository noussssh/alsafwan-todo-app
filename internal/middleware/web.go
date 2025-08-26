package middleware

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"alsafwanmarine.com/todo-app/internal/models"
	"github.com/gin-gonic/gin"
)

type WebMiddleware struct {
	templates *template.Template
}

func NewWebMiddleware() *WebMiddleware {
	return &WebMiddleware{}
}

func (w *WebMiddleware) LoadTemplates(templateDir string) error {
	templates, err := template.ParseGlob(templateDir + "/**/*.html")
	if err != nil {
		return err
	}
	w.templates = templates
	return nil
}

func (w *WebMiddleware) TemplateRenderer() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("templates", w.templates)
		c.Next()
	}
}

type PageData struct {
	Title         string
	User          *models.User
	ActiveNav     string
	FlashSuccess  string
	FlashError    string
	FlashWarning  string
	FlashInfo     string
	Errors        map[string]string
	FormData      map[string]interface{}
}

func (w *WebMiddleware) FlashMessages() gin.HandlerFunc {
	return func(c *gin.Context) {
		if flashSuccess, err := c.Cookie("flash_success"); err == nil {
			c.Set("flash_success", flashSuccess)
			c.SetCookie("flash_success", "", -1, "/", "", false, true)
		}

		if flashError, err := c.Cookie("flash_error"); err == nil {
			c.Set("flash_error", flashError)
			c.SetCookie("flash_error", "", -1, "/", "", false, true)
		}

		if flashWarning, err := c.Cookie("flash_warning"); err == nil {
			c.Set("flash_warning", flashWarning)
			c.SetCookie("flash_warning", "", -1, "/", "", false, true)
		}

		if flashInfo, err := c.Cookie("flash_info"); err == nil {
			c.Set("flash_info", flashInfo)
			c.SetCookie("flash_info", "", -1, "/", "", false, true)
		}

		c.Next()
	}
}

func RenderPage(c *gin.Context, templateName string, title string, data interface{}) {
	_, exists := c.Get("templates")
	if !exists {
		c.String(http.StatusInternalServerError, "Templates not loaded")
		return
	}

	// Create page data with common fields
	pageData := PageData{
		Title:     title,
		ActiveNav: c.GetString("active_nav"),
		Errors:    make(map[string]string),
		FormData:  make(map[string]interface{}),
	}

	// Add current user
	if user := GetCurrentUser(c); user != nil {
		pageData.User = user
	}

	// Add flash messages
	if flash, exists := c.Get("flash_success"); exists {
		pageData.FlashSuccess = flash.(string)
	}
	if flash, exists := c.Get("flash_error"); exists {
		pageData.FlashError = flash.(string)
	}
	if flash, exists := c.Get("flash_warning"); exists {
		pageData.FlashWarning = flash.(string)
	}
	if flash, exists := c.Get("flash_info"); exists {
		pageData.FlashInfo = flash.(string)
	}

	// Add custom data
	if data != nil {
		c.Set("page_data", data)
	}

	c.Set("page_info", pageData)
	c.HTML(http.StatusOK, "layouts/base.html", gin.H{
		"Title":       pageData.Title,
		"User":        pageData.User,
		"ActiveNav":   pageData.ActiveNav,
		"FlashSuccess": pageData.FlashSuccess,
		"FlashError":  pageData.FlashError,
		"FlashWarning": pageData.FlashWarning,
		"FlashInfo":   pageData.FlashInfo,
		"Data":        data,
	})
}

func SetFlash(c *gin.Context, flashType, message string) {
	c.SetCookie(
		"flash_"+flashType,
		message,
		300, // 5 minutes
		"/",
		"",
		false,
		true,
	)
}

func SetFlashSuccess(c *gin.Context, message string) {
	SetFlash(c, "success", message)
}

func SetFlashError(c *gin.Context, message string) {
	SetFlash(c, "error", message)
}

func SetFlashWarning(c *gin.Context, message string) {
	SetFlash(c, "warning", message)
}

func SetFlashInfo(c *gin.Context, message string) {
	SetFlash(c, "info", message)
}

func RequireWebAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireWebRole(minRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if int(user.Role) > int(minRole) {
			SetFlashError(c, "Access denied. Insufficient permissions.")
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		c.Next()
	}
}

func ParseFormErrors(c *gin.Context, err error) map[string]string {
	errors := make(map[string]string)
	
	if err != nil {
		// Handle validation errors
		if strings.Contains(err.Error(), "email") {
			errors["Email"] = "Please enter a valid email address"
		}
		if strings.Contains(err.Error(), "required") {
			errors["General"] = "Please fill in all required fields"
		}
		// Add more specific error parsing as needed
	}
	
	return errors
}

func GetFormData(c *gin.Context) map[string]interface{} {
	data := make(map[string]interface{})
	
	// Parse form data from request
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		c.Request.ParseForm()
		for key, values := range c.Request.Form {
			if len(values) > 0 {
				// Handle checkboxes and boolean values
				if values[0] == "true" || values[0] == "false" {
					data[key] = values[0] == "true"
				} else if i, err := strconv.Atoi(values[0]); err == nil {
					data[key] = i
				} else {
					data[key] = values[0]
				}
			}
		}
	}
	
	return data
}

func SetActiveNav(nav string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("active_nav", nav)
		c.Next()
	}
}