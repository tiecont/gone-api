package controllers

import (
	"github.com/gin-gonic/gin"
	"gone-be/src/modules/role/services"
	"net/http"
)

func CreateRole(c *gin.Context) {
	var request struct {
		RoleName string `json:"name"`
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

	results, err := services.CreateRoleByAdmin(request.RoleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": gin.H{
				"code":    http.StatusInternalServerError,
				"message": err.Error(),
			},
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status": gin.H{
			"code":    http.StatusCreated,
			"message": "Role Created Successfully",
		},
		"data": results,
	})
}

func ListRoles(c *gin.Context) {
	roles, err := services.ListAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": gin.H{
				"code":    http.StatusInternalServerError,
				"message": err.Error(),
			},
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Successfully List All Roles",
		},
		"data": roles,
	})
}
