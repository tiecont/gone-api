package controllers

import (
	"bytes"
	model "gone-be/src/modules/user/models"
	"gone-be/src/modules/user/services"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) {
	// Retrieve the user from the context (set by Authorization middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Unauthorized: user not found",
			},
		})
		return
	}

	// Type assertion (since c.Get returns an interface{})
	currentUser, _ := user.(*model.User)

	// Fetch user info from service
	users, err := services.GetUserInfo(currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Failed to get user",
			},
			"error": err.Error(),
		})
		return
	}

	// Return user data
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Retrieved the profile of user successfully",
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
		Email       string `json:"email"`
		Username    string `json:"username"`
		PhoneNumber string `json:"phone_number"`
		Password    string `json:"password"`
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
	response, status, err := services.LoginUser(request.Email, request.PhoneNumber, request.Password, request.Username)
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
			"message": "Verification code sent to your email. Please verify to process.",
		},
		"data": response,
	})
}

func LoginGoogle(c *gin.Context) {
	var request struct {
		Token string `json:"token"`
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
	response, err := services.LoginGoogle(request.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": gin.H{
				"code":    http.StatusBadRequest,
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

func LoginTwitter(c *gin.Context) {
	// Read and log the raw request body
	body, _ := io.ReadAll(c.Request.Body)
	// Reset the request body for further processing
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// Define the request struct
	var request struct {
		Username               string `json:"username"`
		Email                  string `json:"email"`
		Image                  string `json:"image"`
		ProfileBackgroundImage string `json:"background_image"`
		ProfileBackgroundColor string `json:"background_color"`
		TwitterID              string `json:"twitter_id"`
	}

	// Bind JSON to the struct
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Println("Error binding JSON:", err)
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
	response, err := services.LoginTwitter(
		request.Username,
		request.Email,
		request.Image,
		request.ProfileBackgroundImage,
		request.ProfileBackgroundColor,
		request.TwitterID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": gin.H{
				"code":    http.StatusBadRequest,
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

// Refresh token

func RefreshAccessToken(c *gin.Context) {
	// Step 1: Get the refresh-token from header
	authHeader := c.GetHeader("x-rtoken-id")
	clientID := c.GetHeader("x-client-id")

	// Return error
	if authHeader == "" || clientID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": gin.H{
				"code":    http.StatusUnauthorized,
				"message": "missing token or client id on header",
			},
		})
		return
	}

	// Step 2: Extract the Bearer token
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": gin.H{
				"code":    http.StatusUnauthorized,
				"message": "invalid token header format",
			}})
		return
	}
	accessToken := tokenParts[1]

	// Call the service to verify the code and generate tokens
	response, err := services.RefreshAccessToken(accessToken, clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Failed to get user",
			},
			"error": err.Error(),
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Verification successful",
		},
		"data": response,
	})
}

func ForgotPassword(c *gin.Context) {
	// Read and log the raw request body
	body, _ := io.ReadAll(c.Request.Body)
	// Reset the request body for further processing
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	var request struct {
		Email string `json:"email"`
	}
	// Bind JSON to the struct
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status": gin.H{
				"code":    http.StatusBadRequest,
				"message": "Invalid input",
			},
			"error": err.Error(),
		})
		return
	}

	var status, err = services.ForgotPassword(request.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": gin.H{
				"code":    status,
				"message": "Forbidden",
			},
			"error": err.Error(),
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Verification email had been sent",
		},
	})
}

func ChangePassword(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	var request struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
		UserID      uint   `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": gin.H{
				"code":    http.StatusBadRequest,
				"message": "Invalid input",
			},
		})
		return
	}
	status, err := services.ChangePassword(request.OldPassword, request.NewPassword, strconv.Itoa(int(request.UserID)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": gin.H{
				"code":    status,
				"message": "Failed to change password",
			},
		})
		return
	}
	// Return results
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    status,
			"message": "Successfully changed password",
		},
	})
}
