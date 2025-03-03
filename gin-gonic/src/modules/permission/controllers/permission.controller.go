package controllers

import (
	"github.com/gin-gonic/gin"
	"gone-be/src/modules/permission/services"
	"net/http"
)

func CreatePermission(c *gin.Context) {
	var request struct {
		PermissionName string `json:"name"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	isCreated, err := services.CreatePermission(request.PermissionName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !isCreated {
		c.JSON(http.StatusBadRequest, gin.H{"error": "An error occurred during creating permission"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status": gin.H{
			"code":    http.StatusCreated,
			"message": "Permission Created Successfully",
		},
	})
}

func ListPermissions(c *gin.Context) {
	results, err := services.ListAllPermissions()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Successfully Get All Permissions",
		},
		"data": results,
	})
}

func AssignPermissionsToRole(c *gin.Context) {
	var request struct {
		RoleID        uint   `json:"role_id"`
		PermissionIDs []uint `json:"permission_ids"`
	}

	// Validate request payload
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call service function
	results, err := services.AddPermissionsToRole(request.RoleID, request.PermissionIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Permissions Assigned Successfully",
		},
		"data": results,
	})
}
