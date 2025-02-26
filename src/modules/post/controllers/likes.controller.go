package controllers

import (
	"github.com/gin-gonic/gin"
	"gone-be/src/functions"
	"gone-be/src/modules/post/services"
	"gone-be/src/utils"
	"net/http"
)

func Like(c *gin.Context) {
	var request struct {
		PostID uint `json:"post_id"`
	}
	err := utils.BindJson(c, request)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	// Get current user from context
	currentUser, getUserErr := functions.GetCurrentUser(c)
	if getUserErr != nil {
		c.JSON(getUserErr.StatusCode, gin.H{"error": getUserErr.Message})
		return
	}
	err = services.Like(currentUser.ID, request.PostID)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Successfully liked post",
		},
	})
}
