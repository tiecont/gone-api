package controllers

import (
	"github.com/gin-gonic/gin"
	"gone-be/src/modules/post/services"
	model "gone-be/src/modules/user/models"
	"net/http"
	"strings"
)

// CreatePost handles creating a new post
func CreatePost(c *gin.Context) {
	// Step 1: Parse form data
	title := strings.TrimSpace(c.PostForm("title"))
	content := strings.TrimSpace(c.PostForm("content"))
	categoryIds := c.PostFormArray("category_ids") // Get array of category IDs

	if len(categoryIds) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one category must be selected"})
		return
	}

	// Step 2: Handle multiple file uploads
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}
	// Handle Receive multiple media
	files := form.File["media"]

	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one image file is required"})
		return
	}

	// Step 3: Get user
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
	// Step 4: Call the service to create a post
	result, err := services.CreatePost(title, content, categoryIds, files, currentUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Step 5: Return success response
	c.JSON(http.StatusCreated, gin.H{
		"status": gin.H{
			"code":    http.StatusCreated,
			"message": "Post Created Successfully",
		},
		"data": result,
	})
}

func ListAllPost(c *gin.Context) {
	var request struct {
		CurrentPage  int    `json:"current_page"`
		ItemsPerPage int    `json:"items_per_page"`
		OrderBy      string `json:"order_by"`
		SortBy       string `json:"sort_by"`
		SearchValue  string `json:"search_value"`
	}
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": gin.H{
				"code":    http.StatusBadRequest,
				"message": "Invalid input",
			},
			"error": err.Error(),
		})
		return
	}
	response, err := services.List(request.CurrentPage, request.ItemsPerPage, request.OrderBy, request.SortBy, request.SearchValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "List all posts successfully",
		},
		"data": response,
	})
}

func GetDetailsPost(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id is required"})
		return
	}
	response, err := services.GetDetails(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"code":    http.StatusOK,
			"message": "Successfully get details of this post",
		},
		"data": response,
	})
}
