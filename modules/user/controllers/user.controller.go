package controllers

import (
	model "gone-be/modules/user/models"
	"gone-be/modules/user/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	// Chỉ gọi service để lấy dữ liệu
	users, err := services.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users from the database"})
		return
	}

	// Trả về danh sách người dùng
	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	// Step 1: Parse JSON request body into `user_data`
	var userData model.User
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": gin.H{
				"code":    http.StatusBadRequest,
				"message": "Invalid input",
			},
			"error": err.Error(),
		})
		return
	}

	// Step 2: Call the service layer to handle user creation
	response, err := services.HandleCreateUser(userData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Failed to create user",
			},
			"error": err.Error(),
		})
		return
	}

	// Step 3: Return a success response
	c.JSON(http.StatusCreated, gin.H{
		"status": gin.H{
			"code":    http.StatusCreated,
			"message": "User has been created successfully",
		},
		"data": response,
	})
}
