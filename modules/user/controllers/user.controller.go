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
	var user model.User

	// Nhận dữ liệu từ request và gửi đến service
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Gửi dữ liệu đến service để xử lý
	response, err := services.HandleCreateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trả về response từ service
	c.JSON(http.StatusCreated, response)
}
