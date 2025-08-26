package controllers

import (
	"net/http"
	"strconv"

	"alsafwanmarine.com/todo-app/internal/middleware"
	"alsafwanmarine.com/todo-app/internal/models"
	"alsafwanmarine.com/todo-app/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	db                   *gorm.DB
	activityService      *services.ActivityService
	passwordResetService *services.PasswordResetService
}

func NewUserController(db *gorm.DB, activityService *services.ActivityService, passwordResetService *services.PasswordResetService) *UserController {
	return &UserController{
		db:                   db,
		activityService:      activityService,
		passwordResetService: passwordResetService,
	}
}

func (uc *UserController) ListUsers(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	var users []models.User
	query := uc.db
	
	if currentUser.Role == models.RoleManager {
		query = query.Where("role = ?", models.RoleSalesperson)
	}
	
	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (uc *UserController) GetUser(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var user models.User
	if err := uc.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}
	
	if !currentUser.CanManageUser(&user) && currentUser.ID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"user": user})
}

type CreateUserRequest struct {
	Email    string            `json:"email" binding:"required,email"`
	Name     string            `json:"name" binding:"required"`
	Role     models.UserRole   `json:"role" binding:"required"`
	Company  *string           `json:"company"`
	Password string            `json:"password" binding:"required"`
}

func (uc *UserController) CreateUser(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if currentUser.Role == models.RoleManager && req.Role != models.RoleSalesperson {
		c.JSON(http.StatusForbidden, gin.H{"error": "Managers can only create salespeople"})
		return
	}
	
	if err := models.ValidateName(req.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := models.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := models.ValidateCompany(req.Company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	user := models.User{
		Email:   req.Email,
		Name:    req.Name,
		Role:    req.Role,
		Company: req.Company,
		Enabled: true,
	}
	
	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set password"})
		return
	}
	
	if err := uc.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	
	uc.activityService.LogUserCRUD(currentUser, &user, "create", c.ClientIP(), c.Request.UserAgent())
	
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

type UpdateUserRequest struct {
	Email   *string          `json:"email"`
	Name    *string          `json:"name"`
	Role    *models.UserRole `json:"role"`
	Company *string          `json:"company"`
	Enabled *bool            `json:"enabled"`
}

func (uc *UserController) UpdateUser(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var user models.User
	if err := uc.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}
	
	if !currentUser.CanManageUser(&user) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if req.Role != nil && currentUser.Role == models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "Managers cannot change user roles"})
		return
	}
	
	if req.Enabled != nil && !currentUser.CanDisableUser(&user) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot disable this user"})
		return
	}
	
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Name != nil {
		if err := models.ValidateName(*req.Name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user.Name = *req.Name
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Company != nil {
		if err := models.ValidateCompany(req.Company); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user.Company = req.Company
	}
	if req.Enabled != nil {
		user.Enabled = *req.Enabled
	}
	
	if err := uc.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	
	uc.activityService.LogUserCRUD(currentUser, &user, "update", c.ClientIP(), c.Request.UserAgent())
	
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (uc *UserController) DeleteUser(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var user models.User
	if err := uc.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}
	
	if !currentUser.CanManageUser(&user) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	
	if currentUser.ID == user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete your own account"})
		return
	}
	
	if err := uc.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	
	uc.activityService.LogUserCRUD(currentUser, &user, "delete", c.ClientIP(), c.Request.UserAgent())
	
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

type AdminResetPasswordRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (uc *UserController) ResetPassword(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var user models.User
	if err := uc.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}
	
	if !currentUser.CanManageUser(&user) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	
	var req AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	newPassword, err := uc.passwordResetService.ManualReset(uint(userID), currentUser.ID, req.Reason, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
		"new_password": newPassword,
	})
}

func (uc *UserController) ToggleEnabled(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var user models.User
	if err := uc.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		}
		return
	}
	
	if !currentUser.CanDisableUser(&user) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot disable this user"})
		return
	}
	
	user.Enabled = !user.Enabled
	if err := uc.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}
	
	action := "enable"
	if !user.Enabled {
		action = "disable"
	}
	
	uc.activityService.LogUserCRUD(currentUser, &user, action, c.ClientIP(), c.Request.UserAgent())
	
	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"message": "User status updated successfully",
	})
}

type BulkResetPasswordsRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required"`
	Reason  string `json:"reason" binding:"required"`
}

func (uc *UserController) BulkResetPasswords(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	var req BulkResetPasswordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	results, err := uc.passwordResetService.BulkResetPasswords(req.UserIDs, currentUser.ID, req.Reason, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset passwords"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Bulk password reset completed",
		"results": results,
	})
}

type BulkToggleEnabledRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required"`
	Enabled bool   `json:"enabled"`
}

func (uc *UserController) BulkToggleEnabled(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	var req BulkToggleEnabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var users []models.User
	if err := uc.db.Where("id IN ?", req.UserIDs).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	
	updatedUsers := []models.User{}
	for _, user := range users {
		if currentUser.CanDisableUser(&user) {
			user.Enabled = req.Enabled
			if err := uc.db.Save(&user).Error; err == nil {
				updatedUsers = append(updatedUsers, user)
				
				action := "enable"
				if !req.Enabled {
					action = "disable"
				}
				uc.activityService.LogUserCRUD(currentUser, &user, action, c.ClientIP(), c.Request.UserAgent())
			}
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Bulk status update completed",
		"updated_count": len(updatedUsers),
		"users": updatedUsers,
	})
}