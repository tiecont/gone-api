package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	// Trả về danh sách người dùng (giả sử đây là dữ liệu mẫu)
	users := []map[string]string{
		{"id": "1", "name": "John Doe"},
		{"id": "2", "name": "Jane Doe"},
	}
	c.JSON(http.StatusOK, users)
}

// POST: Register User
func CreateUser(c *gin.Context) {

}
