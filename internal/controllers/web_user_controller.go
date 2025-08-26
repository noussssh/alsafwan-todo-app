package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"alsafwanmarine.com/todo-app/internal/middleware"
	"alsafwanmarine.com/todo-app/internal/models"
	"alsafwanmarine.com/todo-app/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WebUserController struct {
	db                   *gorm.DB
	activityService      *services.ActivityService
	passwordResetService *services.PasswordResetService
}

func NewWebUserController(db *gorm.DB, activityService *services.ActivityService, passwordResetService *services.PasswordResetService) *WebUserController {
	return &WebUserController{
		db:                   db,
		activityService:      activityService,
		passwordResetService: passwordResetService,
	}
}

func (uc *WebUserController) ListUsers(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get search/filter parameters with pagination
	searchQuery := c.Query("search")
	filterRole := c.Query("role")
	filterStatus := c.Query("status")
	
	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := 20 // Users per page
	offset := (page - 1) * limit

	var users []models.User
	var totalUsers int64
	
	// Start with base query and count
	query := uc.db.Model(&models.User{})
	countQuery := uc.db.Model(&models.User{})

	// Apply role-based filtering
	if currentUser.Role == models.RoleManager {
		query = query.Where("role = ?", models.RoleSalesperson)
		countQuery = countQuery.Where("role = ?", models.RoleSalesperson)
	}

	// Apply search filter (optimized with index)
	if searchQuery != "" {
		searchCondition := "name LIKE ? OR email LIKE ?"
		searchParam := "%" + searchQuery + "%"
		query = query.Where(searchCondition, searchParam, searchParam)
		countQuery = countQuery.Where(searchCondition, searchParam, searchParam)
	}

	// Apply role filter
	if filterRole != "" {
		if role, err := strconv.Atoi(filterRole); err == nil {
			query = query.Where("role = ?", role)
			countQuery = countQuery.Where("role = ?", role)
		}
	}

	// Apply status filter (uses index)
	if filterStatus == "enabled" {
		query = query.Where("enabled = ?", true)
		countQuery = countQuery.Where("enabled = ?", true)
	} else if filterStatus == "disabled" {
		query = query.Where("enabled = ?", false)
		countQuery = countQuery.Where("enabled = ?", false)
	}

	// Get total count for pagination
	if err := countQuery.Count(&totalUsers).Error; err != nil {
		middleware.SetFlashError(c, "Failed to count users")
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Get users with pagination, select only needed fields for list view
	if err := query.
		Select("id, name, email, role, company, enabled, created_at, last_sign_in_at").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		middleware.SetFlashError(c, "Failed to load users")
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Calculate pagination info
	totalPages := int((totalUsers + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrev := page > 1

	data := gin.H{
		"Title":        "User Management",
		"User":         currentUser,
		"ActiveNav":    "users",
		"Users":        users,
		"SearchQuery":  searchQuery,
		"FilterRole":   filterRole,
		"FilterStatus": filterStatus,
		"Pagination": gin.H{
			"CurrentPage": page,
			"TotalPages":  totalPages,
			"TotalUsers":  totalUsers,
			"HasNext":     hasNext,
			"HasPrev":     hasPrev,
			"NextPage":    page + 1,
			"PrevPage":    page - 1,
		},
	}

	c.HTML(http.StatusOK, "base.html", data)
}

func (uc *WebUserController) ShowUser(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.SetFlashError(c, "Invalid user ID")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	var viewUser models.User
	if err := uc.db.First(&viewUser, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.SetFlashError(c, "User not found")
		} else {
			middleware.SetFlashError(c, "Failed to load user")
		}
		c.Redirect(http.StatusFound, "/users")
		return
	}

	// Check permissions
	if !currentUser.CanManageUser(&viewUser) && currentUser.ID != viewUser.ID {
		middleware.SetFlashError(c, "Access denied")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	// Get user activities
	userActivities, _ := uc.activityService.GetUserActivities(viewUser.ID, 20)

	// Get password reset events (if allowed)
	var passwordResets []models.PasswordResetEvent
	if currentUser.CanManageUser(&viewUser) {
		passwordResets, _ = uc.passwordResetService.GetResetEvents(viewUser.ID)
	}

	data := gin.H{
		"Title":          "User Details",
		"User":           currentUser,
		"ActiveNav":      "users",
		"ViewUser":       viewUser,
		"UserActivities": userActivities,
		"PasswordResets": passwordResets,
	}

	c.HTML(http.StatusOK, "base.html", data)
}

func (uc *WebUserController) ShowCreateUser(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	data := gin.H{
		"Title":    "Create User",
		"User":     currentUser,
		"ActiveNav": "users",
		"IsEdit":   false,
		"Errors":   make(map[string]string),
		"FormData": make(map[string]interface{}),
	}

	c.HTML(http.StatusOK, "base.html", data)
}

func (uc *WebUserController) HandleCreateUser(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Parse form data
	name := strings.TrimSpace(c.PostForm("name"))
	email := strings.TrimSpace(strings.ToLower(c.PostForm("email")))
	roleStr := c.PostForm("role")
	company := c.PostForm("company")
	password := c.PostForm("password")
	passwordConfirm := c.PostForm("password_confirm")
	enabled := c.PostForm("enabled") == "true"

	// Validate form data
	errors := make(map[string]string)
	formData := gin.H{
		"Name":     name,
		"Email":    email,
		"Role":     roleStr,
		"Company":  company,
		"Enabled":  enabled,
	}

	if err := models.ValidateName(name); err != nil {
		errors["Name"] = err.Error()
	}

	if email == "" {
		errors["Email"] = "Email is required"
	}

	role, err := strconv.Atoi(roleStr)
	if err != nil || role < 0 || role > 2 {
		errors["Role"] = "Please select a valid role"
	}

	// Check manager permissions
	if currentUser.Role == models.RoleManager && models.UserRole(role) != models.RoleSalesperson {
		errors["Role"] = "Managers can only create salespeople"
	}

	if company != "" {
		if err := models.ValidateCompany(&company); err != nil {
			errors["Company"] = err.Error()
		}
	}

	if err := models.ValidatePassword(password); err != nil {
		errors["Password"] = err.Error()
	}

	if password != passwordConfirm {
		errors["PasswordConfirm"] = "Passwords do not match"
	}

	// Check for existing email
	var existingUser models.User
	if uc.db.Where("email = ?", email).First(&existingUser).Error == nil {
		errors["Email"] = "Email address is already in use"
	}

	if len(errors) > 0 {
		data := gin.H{
			"Title":    "Create User",
			"User":     currentUser,
			"ActiveNav": "users",
			"IsEdit":   false,
			"Errors":   errors,
			"FormData": formData,
		}
		c.HTML(http.StatusBadRequest, "base.html", data)
		return
	}

	// Create user
	user := models.User{
		Name:    name,
		Email:   email,
		Role:    models.UserRole(role),
		Enabled: enabled,
	}

	if company != "" {
		user.Company = &company
	}

	if err := user.SetPassword(password); err != nil {
		errors["General"] = "Failed to set password"
		data := gin.H{
			"Title":    "Create User",
			"User":     currentUser,
			"ActiveNav": "users",
			"IsEdit":   false,
			"Errors":   errors,
			"FormData": formData,
		}
		c.HTML(http.StatusInternalServerError, "base.html", data)
		return
	}

	if err := uc.db.Create(&user).Error; err != nil {
		errors["General"] = "Failed to create user"
		data := gin.H{
			"Title":    "Create User",
			"User":     currentUser,
			"ActiveNav": "users",
			"IsEdit":   false,
			"Errors":   errors,
			"FormData": formData,
		}
		c.HTML(http.StatusInternalServerError, "base.html", data)
		return
	}

	// Log activity
	uc.activityService.LogUserCRUD(currentUser, &user, "create", c.ClientIP(), c.Request.UserAgent())

	middleware.SetFlashSuccess(c, "User created successfully!")
	c.Redirect(http.StatusFound, "/users/"+strconv.Itoa(int(user.ID)))
}

func (uc *WebUserController) ShowEditUser(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.SetFlashError(c, "Invalid user ID")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	var editUser models.User
	if err := uc.db.First(&editUser, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.SetFlashError(c, "User not found")
		} else {
			middleware.SetFlashError(c, "Failed to load user")
		}
		c.Redirect(http.StatusFound, "/users")
		return
	}

	// Check permissions
	if !currentUser.CanManageUser(&editUser) {
		middleware.SetFlashError(c, "Access denied")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	data := gin.H{
		"Title":    "Edit User",
		"User":     currentUser,
		"ActiveNav": "users",
		"IsEdit":   true,
		"EditUser": editUser,
		"Errors":   make(map[string]string),
	}

	c.HTML(http.StatusOK, "base.html", data)
}

func (uc *WebUserController) HandleEditUser(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.SetFlashError(c, "Invalid user ID")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	var editUser models.User
	if err := uc.db.First(&editUser, userID).Error; err != nil {
		middleware.SetFlashError(c, "User not found")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	// Check permissions
	if !currentUser.CanManageUser(&editUser) {
		middleware.SetFlashError(c, "Access denied")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	// Parse form data
	name := strings.TrimSpace(c.PostForm("name"))
	email := strings.TrimSpace(strings.ToLower(c.PostForm("email")))
	roleStr := c.PostForm("role")
	company := c.PostForm("company")
	enabled := c.PostForm("enabled") == "true"

	// Validate
	errors := make(map[string]string)

	if err := models.ValidateName(name); err != nil {
		errors["Name"] = err.Error()
	}

	if email == "" {
		errors["Email"] = "Email is required"
	}

	role, err := strconv.Atoi(roleStr)
	if err != nil || role < 0 || role > 2 {
		errors["Role"] = "Please select a valid role"
	}

	// Check manager permissions
	if currentUser.Role == models.RoleManager && models.UserRole(role) != models.RoleSalesperson {
		errors["Role"] = "Managers cannot change user roles"
	}

	if company != "" {
		if err := models.ValidateCompany(&company); err != nil {
			errors["Company"] = err.Error()
		}
	}

	// Check for email conflicts
	var existingUser models.User
	if uc.db.Where("email = ? AND id != ?", email, editUser.ID).First(&existingUser).Error == nil {
		errors["Email"] = "Email address is already in use"
	}

	// Check disable permissions
	if !enabled && !currentUser.CanDisableUser(&editUser) {
		errors["Enabled"] = "Cannot disable this user"
		enabled = true // Force enable
	}

	if len(errors) > 0 {
		data := gin.H{
			"Title":    "Edit User",
			"User":     currentUser,
			"ActiveNav": "users",
			"IsEdit":   true,
			"EditUser": editUser,
			"Errors":   errors,
		}
		c.HTML(http.StatusBadRequest, "base.html", data)
		return
	}

	// Update user
	editUser.Name = name
	editUser.Email = email
	editUser.Role = models.UserRole(role)
	editUser.Enabled = enabled

	if company != "" {
		editUser.Company = &company
	} else {
		editUser.Company = nil
	}

	if err := uc.db.Save(&editUser).Error; err != nil {
		errors["General"] = "Failed to update user"
		data := gin.H{
			"Title":    "Edit User",
			"User":     currentUser,
			"ActiveNav": "users",
			"IsEdit":   true,
			"EditUser": editUser,
			"Errors":   errors,
		}
		c.HTML(http.StatusInternalServerError, "base.html", data)
		return
	}

	// Log activity
	uc.activityService.LogUserCRUD(currentUser, &editUser, "update", c.ClientIP(), c.Request.UserAgent())

	middleware.SetFlashSuccess(c, "User updated successfully!")
	c.Redirect(http.StatusFound, "/users/"+strconv.Itoa(int(editUser.ID)))
}

func (uc *WebUserController) HandleDeleteUser(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.SetFlashError(c, "Invalid user ID")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	var deleteUser models.User
	if err := uc.db.First(&deleteUser, userID).Error; err != nil {
		middleware.SetFlashError(c, "User not found")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	// Check permissions
	if !currentUser.CanManageUser(&deleteUser) || currentUser.ID == deleteUser.ID {
		middleware.SetFlashError(c, "Cannot delete this user")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	if err := uc.db.Delete(&deleteUser).Error; err != nil {
		middleware.SetFlashError(c, "Failed to delete user")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	// Log activity
	uc.activityService.LogUserCRUD(currentUser, &deleteUser, "delete", c.ClientIP(), c.Request.UserAgent())

	middleware.SetFlashSuccess(c, "User deleted successfully!")
	c.Redirect(http.StatusFound, "/users")
}

func (uc *WebUserController) HandleToggleStatus(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.SetFlashError(c, "Invalid user ID")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	var targetUser models.User
	if err := uc.db.First(&targetUser, userID).Error; err != nil {
		middleware.SetFlashError(c, "User not found")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	// Check permissions
	if !currentUser.CanDisableUser(&targetUser) {
		middleware.SetFlashError(c, "Cannot modify this user's status")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	// Toggle status
	targetUser.Enabled = !targetUser.Enabled
	if err := uc.db.Save(&targetUser).Error; err != nil {
		middleware.SetFlashError(c, "Failed to update user status")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	// Log activity
	action := "enable"
	if !targetUser.Enabled {
		action = "disable"
	}
	uc.activityService.LogUserCRUD(currentUser, &targetUser, action, c.ClientIP(), c.Request.UserAgent())

	status := "enabled"
	if !targetUser.Enabled {
		status = "disabled"
	}
	middleware.SetFlashSuccess(c, "User "+status+" successfully!")
	c.Redirect(http.StatusFound, "/users/"+strconv.Itoa(int(targetUser.ID)))
}

func (uc *WebUserController) HandleResetPassword(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.SetFlashError(c, "Invalid user ID")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	var targetUser models.User
	if err := uc.db.First(&targetUser, userID).Error; err != nil {
		middleware.SetFlashError(c, "User not found")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	// Check permissions
	if !currentUser.CanManageUser(&targetUser) {
		middleware.SetFlashError(c, "Access denied")
		c.Redirect(http.StatusFound, "/users")
		return
	}

	reason := c.PostForm("reason")
	if reason == "" {
		reason = "Admin initiated password reset"
	}

	newPassword, err := uc.passwordResetService.ManualReset(targetUser.ID, currentUser.ID, reason, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		middleware.SetFlashError(c, "Failed to reset password")
		c.Redirect(http.StatusFound, "/users/"+strconv.Itoa(int(targetUser.ID)))
		return
	}

	middleware.SetFlashSuccess(c, "Password reset successfully! New password: "+newPassword)
	c.Redirect(http.StatusFound, "/users/"+strconv.Itoa(int(targetUser.ID)))
}