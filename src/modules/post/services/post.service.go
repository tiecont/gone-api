package services

import (
	"errors"
	"gone-be/src/config"
	"gone-be/src/modules/post/models"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Allowed image formats
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

// SaveUploadedFile manually saves the uploaded file
func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// CreatePost handles post creation logic
func CreatePost(title string, content string, categoryIDs []string, file *multipart.FileHeader) (map[string]interface{}, error) {
	db := config.DB

	// Validate file type
	fileType := file.Header.Get("Content-Type")
	if !allowedImageTypes[fileType] {
		return nil, errors.New("invalid image format. Allowed: JPEG, PNG, GIF")
	}

	// Generate unique file name
	uniqueFileName := time.Now().Format("20060102150405") + "_" + file.Filename
	imagePath := filepath.Join("uploads", uniqueFileName)

	// Save file manually
	if err := SaveUploadedFile(file, imagePath); err != nil {
		return nil, errors.New("failed to save image")
	}

	// Step 3: Store post, media & categories in database (transaction for consistency)
	tx := db.Begin()
	if tx.Error != nil {
		return nil, errors.New("failed to start transaction")
	}

	// Create post
	post := models.Post{
		Title:   title,
		Content: content,
	}
	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create post")
	}

	// Create media entry
	media := models.Media{
		PostID: post.ID,
		Path:   imagePath,
		Type:   fileType,
	}
	if err := tx.Create(&media).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to save media")
	}

	// Convert string categoryID to int inside the loop
	var postCategories []models.PostCategory
	for _, categoryID := range categoryIDs {
		// Convert string to int
		id, err := strconv.Atoi(categoryID)
		if err != nil {
			// Handle the error, for example:
			return nil, errors.New("invalid category ID")
		}

		// Append to postCategories
		postCategories = append(postCategories, models.PostCategory{
			PostID:     post.ID,
			CategoryID: uint(id),
		})
	}

	if len(postCategories) > 0 {
		if err := tx.Create(&postCategories).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("failed to associate categories")
		}
	}

	// Commit transaction
	tx.Commit()

	// Return result
	return map[string]interface{}{
		"id":      post.ID,
		"title":   post.Title,
		"content": post.Content,
		"media": map[string]interface{}{
			"path": media.Path,
			"type": media.Type,
		},
		"categories": categoryIDs,
	}, nil
}
