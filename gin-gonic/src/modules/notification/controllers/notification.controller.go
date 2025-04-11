package controllers

import (
	"github.com/gin-gonic/gin"
	"gone-be/src/functions"
	"gone-be/src/modules/notification/services"
	"gone-be/src/utils"
	"net/http"
)

func List(c *gin.Context) {
	var request struct {
		CurrentPage  int    `json:"current_page"`
		ItemsPerPage int    `json:"items_per_page"`
		OrderBy      string `json:"order_by"`
		SortBy       string `json:"sort_by"`
		SearchValue  string `json:"search_value"`
		IsSeen       bool   `json:"is_seen"`
		All          bool   `json:"all"`
	}
	bindErr := utils.BindJson(c, &request)
	if bindErr != nil {
		c.JSON(bindErr.StatusCode, gin.H{"error": bindErr.Message})
		return
	}

	// Get current user from context
	currentUser, getUserErr := functions.GetCurrentUser(c)
	if getUserErr != nil {
		c.JSON(getUserErr.StatusCode, gin.H{"error": getUserErr.Message})
		return
	}
	response, err := services.GetNotificationsByUserID(currentUser.ID, request.CurrentPage, request.ItemsPerPage, request.OrderBy, request.SortBy, request.SearchValue, request.IsSeen, request.All)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "List all notifications successfully",
		},
		"data": response,
	})
}

func UpdateSeen(c *gin.Context) {
	var request struct {
		NotificationID uint `json:"notification_id"`
	}
	bindErr := utils.BindJson(c, &request)
	if bindErr != nil {
		c.JSON(bindErr.StatusCode, gin.H{"error": bindErr.Message})
		return
	}
	// Get current user from context
	currentUser, getUserErr := functions.GetCurrentUser(c)
	if getUserErr != nil {
		c.JSON(getUserErr.StatusCode, gin.H{"error": getUserErr.Message})
		return
	}

	serviceErr := services.UpdateIsSeen(request.NotificationID, currentUser.ID)
	if serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Update notification status successfully",
		},
	})
}

// UpdateSeenWithUserID is a controller receive userID, update all notifications for this user
func UpdateSeenWithUserID(c *gin.Context) {
	// Get current user from context
	currentUser, getUserErr := functions.GetCurrentUser(c)
	if getUserErr != nil {
		c.JSON(getUserErr.StatusCode, gin.H{"error": getUserErr.Message})
		return
	}

	serviceErr := services.UpdateIsSeenForAllNotificationByUserID(currentUser.ID)
	if serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Update notifications status successfully",
		},
	})
}
