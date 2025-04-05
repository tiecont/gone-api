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
	totalLikes, err := services.Like(currentUser.ID, request.PostID)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Successfully liked post",
		},
		"data": gin.H{
			"totalLikes": totalLikes,
		},
	})
}

func DisLike(c *gin.Context) {
	var request struct {
		PostID uint `json:"post_id"`
	}
	bindErr := utils.BindJson(c, &request)
	if bindErr != nil {
		c.JSON(bindErr.StatusCode, gin.H{"error": bindErr.Message})
		return
	}
	currentUser, getUserErr := functions.GetCurrentUser(c)
	if getUserErr != nil {
		c.JSON(getUserErr.StatusCode, gin.H{"error": getUserErr.Message})
		return
	}
	totalLikes, err := services.DisLike(currentUser.ID, request.PostID)
	if err != nil {
		c.JSON(err.StatusCode, gin.H{"error": err.Message})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Successfully disliked post",
		},
		"data": gin.H{
			"totalLikes": totalLikes,
		},
	})
}
