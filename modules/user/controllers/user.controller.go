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
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Failed to create user",
			},
			"error": err.Error(),
		})
		return
	}
	// Trả về danh sách người dùng
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Retrieved the list of users successfully",
		},
		"data": users,
	})

}

func RegisterUser(c *gin.Context) {
	// Step 1: Parse JSON request body into `

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
	response, err := services.RegisterUser(userData)
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

func LoginUser(c *gin.Context) {
	// Define a struct to parse the incoming JSON request body
	var request struct {
		Email         string `json:"email"`
		PhoneNumber   string `json:"phone_number"`
		Password      string `json:"password"`
		FacebookToken string `json:"facebook_token"`
		GoogleToken   string `json:"google_token"`
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

	// Call the LoginUser service
	response, status, err := services.LoginUser(request.Email, request.PhoneNumber, request.Password, request.GoogleToken, request.FacebookToken)
	if err != nil {
		c.JSON(status, gin.H{
			"status": gin.H{
				"code":    status,
				"message": err.Error(),
			},
		})
		return
	}

	// Return success response with the tokens
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Login successful",
		},
		"data": response,
	})
}
