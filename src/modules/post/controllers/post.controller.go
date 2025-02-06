package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gone-be/src/modules/post/services"
	"net/http"
	"strings"
)

// CreatePost handles creating a new post
func CreatePost(c *gin.Context) {
	fmt.Printf(c.PostForm("title"))
	fmt.Printf(c.PostForm("content"))
	fmt.Printf(c.PostForm("category_ids"))
	// Step 1: Parse form data
	title := strings.TrimSpace(c.PostForm("title"))
	content := strings.TrimSpace(c.PostForm("content"))
	categoryIds := c.PostFormArray("category_ids") // Get array of category IDs

	fmt.Println(categoryIds) // []
	if len(categoryIds) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one category must be selected"})
		return
	}

	// Step 2: Handle file upload
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}

	// Call the service to create a post
	result, err := services.CreatePost(title, content, categoryIds, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Step 4: Return success response
	c.JSON(http.StatusCreated, gin.H{
		"status": gin.H{
			"code":    http.StatusCreated,
			"message": "Post Created Successfully",
		},
		"data": result,
	})
}
