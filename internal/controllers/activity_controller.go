package controllers

import (
	"net/http"
	"strconv"

	"alsafwanmarine.com/todo-app/internal/middleware"
	"alsafwanmarine.com/todo-app/internal/models"
	"alsafwanmarine.com/todo-app/internal/services"
	"github.com/gin-gonic/gin"
)

type ActivityController struct {
	activityService *services.ActivityService
}

func NewActivityController(activityService *services.ActivityService) *ActivityController {
	return &ActivityController{
		activityService: activityService,
	}
}

func (ac *ActivityController) GetAllActivities(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	if currentUser.Role != models.RoleAdmin && currentUser.Role != models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}
	
	activities, err := ac.activityService.GetAllActivities(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activities"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"activities": activities})
}

func (ac *ActivityController) GetUserActivities(c *gin.Context) {
	currentUser := middleware.GetCurrentUser(c)
	if currentUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	if currentUser.ID != uint(userID) && currentUser.Role != models.RoleAdmin && currentUser.Role != models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}
	
	activities, err := ac.activityService.GetUserActivities(uint(userID), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activities"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"activities": activities})
}