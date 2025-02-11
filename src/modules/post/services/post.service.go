package services

import (
	"database/sql"
	"errors"
	"fmt"
	"gone-be/src/config"
	"gone-be/src/modules/post/models"
	"gone-be/src/utils"
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
func CreatePost(title string, content string, categoryIDs []string, files []*multipart.FileHeader, userID uint) (map[string]interface{}, error) {
	db := config.DB

	// Step 1: Process & Save Multiple Files
	var savedFiles []models.Media

	for _, file := range files {
		// Validate file type
		fileType := file.Header.Get("Content-Type")
		if !allowedImageTypes[fileType] {
			return nil, errors.New("invalid image format. Allowed: JPEG, PNG, GIF")
		}

		// Generate unique file name
		uniqueFileName := time.Now().Format("20060102150405") + "_" + file.Filename
		imagePath := filepath.Join("uploads", uniqueFileName)

		// Save file
		if err := SaveUploadedFile(file, imagePath); err != nil {
			return nil, err
		}

		// Append to saved files
		savedFiles = append(savedFiles, models.Media{
			Path: imagePath,
			Type: fileType,
		})
	}

	// Step 2: Store post & categories in database (transaction for consistency)
	tx := db.Begin()
	if tx.Error != nil {
		return nil, errors.New("failed to start transaction")
	}

	// Create post
	post := models.Post{
		UserID:  userID,
		Title:   title,
		Content: content,
	}
	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create post")
	}

	// Save media records
	for _, media := range savedFiles {
		media.PostID = post.ID
		if err := tx.Create(&media).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("failed to save media")
		}
	}

	// Convert categoryIDs to integers & create relationships
	var postCategories []models.PostCategory
	for _, categoryID := range categoryIDs {
		id, err := strconv.Atoi(categoryID)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("invalid category ID")
		}
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

	// Prepare media response
	var mediaResponse []map[string]interface{}
	for _, media := range savedFiles {
		mediaResponse = append(mediaResponse, map[string]interface{}{
			"path": media.Path,
			"type": media.Type,
		})
	}

	// Return result
	return map[string]interface{}{
		"id":         post.ID,
		"title":      post.Title,
		"content":    post.Content,
		"media":      mediaResponse,
		"categories": categoryIDs,
	}, nil
}

func List(currentPage int, itemsPerPage int, orderBy string, sortBy string, searchValue string) (map[string]interface{}, error) {
	db := config.DB

	// Validate sorting parameters
	if orderBy == "" {
		orderBy = "created_at"
	}
	if sortBy != "asc" && sortBy != "desc" {
		sortBy = "desc"
	}

	// Calculate offset for pagination
	offset := (currentPage - 1) * itemsPerPage
	if offset < 0 {
		offset = 0
	}

	// Phase 1: Get Total Count (for pagination)
	var totalCount int64
	query := db.Model(&models.Post{})
	if searchValue != "" {
		query = query.Where("LOWER(title) LIKE LOWER(?)", "%"+searchValue+"%")
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Phase 2: Fetch Posts with LEFT JOIN on Media Table
	rows, err := db.Table("posts").
		Select(`
			posts.id, posts.title, posts.content, posts.created_at, posts.updated_at,
			media.id AS media_id, media.path AS media_path, media.type AS media_type, media.status AS media_status
		`).
		Joins("LEFT JOIN media ON media.post_id = posts.id").
		Where("LOWER(posts.title) LIKE LOWER(?)", "%"+searchValue+"%").
		Order(fmt.Sprintf("posts.%s %s", orderBy, sortBy)).
		Offset(offset).
		Limit(itemsPerPage).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Phase 3: Process Query Results
	postMap := make(map[uint]map[string]interface{})
	for rows.Next() {
		var postID uint
		var title, content string
		var createdAt, updatedAt time.Time
		var mediaID sql.NullInt64
		var mediaPath, mediaType, mediaStatus sql.NullString

		if err := rows.Scan(
			&postID, &title, &content, &createdAt, &updatedAt,
			&mediaID, &mediaPath, &mediaType, &mediaStatus,
		); err != nil {
			return nil, err
		}

		// If post is not in map, initialize it
		if _, exists := postMap[postID]; !exists {
			postMap[postID] = map[string]interface{}{
				"id":         postID,
				"title":      title,
				"content":    content,
				"created_at": createdAt,
				"updated_at": updatedAt,
				"images":     []map[string]interface{}{}, // Initialize image array
			}
		}

		// Append media if it exists
		if mediaID.Valid {
			postMap[postID]["images"] = append(postMap[postID]["images"].([]map[string]interface{}), map[string]interface{}{
				"id":     mediaID.Int64,
				"path":   mediaPath.String,
				"type":   mediaType.String,
				"status": mediaStatus.String,
			})
		}
	}

	// Convert map to slice
	var items []map[string]interface{}
	for _, post := range postMap {
		items = append(items, post)
	}

	// Phase 4: Generate Pagination Metadata
	paginationResult, err := utils.Paginate(totalCount, currentPage, itemsPerPage)
	if err != nil {
		return nil, err
	}

	// Phase 5: Return Structured Response
	return map[string]interface{}{
		"items":      items,
		"pagination": paginationResult,
	}, nil
}
