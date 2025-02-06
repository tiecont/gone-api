package controllers

import (
	"github.com/gin-gonic/gin"
	"gone-be/src/modules/permission/services"
	"net/http"
)

func CreatePermission(c *gin.Context) {
	var request struct {
		PermissionName string `json:"name"`
		RoleID         uint   `json:"role_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isCreated := services.CreatePermission(request.RoleID, request.PermissionName)
	if !isCreated {
		c.JSON(http.StatusBadRequest, gin.H{"error": "An error occurred during creating permission"})
	}
	c.JSON(http.StatusCreated, gin.H{
		"status": gin.H{
			"code":    http.StatusCreated,
			"message": "Permission Created Successfully",
		},
	})
}

func ListPermissions(c *gin.Context) {
	var request struct {
		RoleID uint `json:"role_id"`
	}
	// Bind the JSON body to the struct
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": gin.H{
				"code":    http.StatusBadRequest,
				"message": "Invalid input",
			},
			"error": err.Error(),
		})
		return
	}
	results, err := services.ListAllPermission(request.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Permissions List Successfully",
		},
		"data": results,
	})
}
