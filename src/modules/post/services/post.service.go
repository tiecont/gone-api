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
	"strings"
	"time"
)

// Allowed Media Types
var allowedMediaTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"video/mp4":  true,
	"video/avi":  true,
	"video/mov":  true,
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

	fmt.Print(categoryIDs)
	// Step 1: Process & Save Media Files (Images & Videos)
	var savedMedia []models.Media

	for _, file := range files {
		// Validate file type
		fileType := file.Header.Get("Content-Type")
		if !allowedMediaTypes[fileType] {
			return nil, errors.New("invalid file format. Allowed: JPEG, PNG, GIF, MP4, AVI, MOV")
		}

		// Generate unique file name
		uniqueFileName := time.Now().Format("20060102150405") + "_" + file.Filename
		mediaPath := filepath.Join("uploads", uniqueFileName)

		// Save file to disk
		if err := SaveUploadedFile(file, mediaPath); err != nil {
			return nil, err
		}

		// Append to saved media list
		savedMedia = append(savedMedia, models.Media{
			Path: mediaPath,
			Type: fileType,
		})
	}

	// Step 2: Store post & categories in database (Transaction)
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

	// Save media records (Images & Videos)
	for _, media := range savedMedia {
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

	// Prepare media response (Separate Images & Videos)
	var imagesResponse []map[string]interface{}
	var videosResponse []map[string]interface{}

	for _, media := range savedMedia {
		mediaItem := map[string]interface{}{
			"id":   media.ID,
			"path": media.Path,
			"type": media.Type,
		}
		if strings.HasPrefix(media.Type, "image/") {
			imagesResponse = append(imagesResponse, mediaItem)
		} else if strings.HasPrefix(media.Type, "video/") {
			videosResponse = append(videosResponse, mediaItem)
		}
	}

	// Return result
	return map[string]interface{}{
		"id":         post.ID,
		"title":      post.Title,
		"content":    post.Content,
		"images":     imagesResponse,
		"videos":     videosResponse,
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

	// **🔹 Fetch Posts with Consolidated Categories & Media**
	rows, err := db.Table("posts").
		Select(`
			posts.id, posts.title, posts.content, posts.created_at, posts.updated_at,
			users.id AS user_id, users.username AS username, profile.profile_pic,
			GROUP_CONCAT(DISTINCT categories.id ORDER BY categories.id ASC SEPARATOR ',') AS category_ids,
			GROUP_CONCAT(DISTINCT categories.name ORDER BY categories.id ASC SEPARATOR ',') AS category_names,
			GROUP_CONCAT(DISTINCT media.id ORDER BY media.id ASC SEPARATOR ',') AS media_ids,
			GROUP_CONCAT(DISTINCT media.path ORDER BY media.id ASC SEPARATOR ',') AS media_paths,
			GROUP_CONCAT(DISTINCT media.type ORDER BY media.id ASC SEPARATOR ',') AS media_types,
			GROUP_CONCAT(DISTINCT media.status ORDER BY media.id ASC SEPARATOR ',') AS media_statuses,
			COUNT(comments.id) AS comment_count, COUNT(post_likes.id) AS likes_count, COUNT(post_shares.id) AS shares_count
		`).
		Joins(`
			LEFT JOIN post_categories ON post_categories.post_id = posts.id
			LEFT JOIN categories ON categories.id = post_categories.category_id
			LEFT JOIN media ON media.post_id = posts.id
			LEFT JOIN users ON users.id = post.user_id
			LEFT JOIN profiles ON profiles.user_id = user.id
			LEFT JOIN comments ON comments.post_id = post.id
			LEFT JOIN post_likes ON post_likes.post_id = post.id
			LEFT JOIN post_shares ON post_shares.post_id = post.id
		`).
		Where("LOWER(posts.content) LIKE LOWER(?)", "%"+searchValue+"%").
		Group("posts.id").
		Order(fmt.Sprintf("posts.%s %s", orderBy, sortBy)).
		Offset(offset).
		Limit(itemsPerPage).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// **🔹 Process Query Results**
	var items []map[string]interface{}
	var totalCount int64

	for rows.Next() {
		var postID, userID uint
		var title, content sql.NullString
		var createdAt, updatedAt time.Time
		var categoryIDs, categoryNames, mediaIDs, mediaPaths, mediaTypes, mediaStatuses sql.NullString
		var username, profilePic string
		var commentsCount
		if err := rows.Scan(
			&postID, &title, &content, &createdAt, &updatedAt,
			&categoryIDs, &categoryNames,
			&mediaIDs, &mediaPaths, &mediaTypes, &mediaStatuses,
			&username, &profilePic, &userID,
		); err != nil {
			return nil, err
		}

		// Parse categories
		var categories []map[string]interface{}
		if categoryIDs.Valid {
			idList := strings.Split(categoryIDs.String, ",")
			nameList := strings.Split(categoryNames.String, ",")
			for i := range idList {
				categories = append(categories, map[string]interface{}{
					"id":   utils.ConvertStringToInt64(idList[i]),
					"name": nameList[i],
				})
			}
		}

		// Parse media
		var images []map[string]interface{}
		var videos []map[string]interface{}
		if mediaIDs.Valid {
			idList := strings.Split(mediaIDs.String, ",")
			pathList := strings.Split(mediaPaths.String, ",")
			typeList := strings.Split(mediaTypes.String, ",")
			statusList := strings.Split(mediaStatuses.String, ",")
			for i := range idList {
				mediaItem := map[string]interface{}{
					"id":     utils.ConvertStringToInt64(idList[i]),
					"path":   pathList[i],
					"type":   typeList[i],
					"status": statusList[i],
				}
				if strings.HasPrefix(typeList[i], "image/") {
					images = append(images, mediaItem)
				} else if strings.HasPrefix(typeList[i], "video/") {
					videos = append(videos, mediaItem)
				}
			}
		}

		// Append post data
		items = append(items, map[string]interface{}{
			"id":         postID,
			"title":      title.String,
			"content":    content.String,
			"categories": categories,
			"created_at": createdAt,
			"updated_at": updatedAt,
			"images":     images,
			"videos":     videos,
		})
		totalCount++
	}

	// **🔹 Generate Pagination Metadata**
	paginationResult, err := utils.Paginate(totalCount, currentPage, itemsPerPage)
	if err != nil {
		return nil, err
	}

	// **🔹 Return Structured Response**
	return map[string]interface{}{
		"items":      items,
		"pagination": paginationResult,
	}, nil
}

func GetDetails(postID string) (map[string]interface{}, error) {
	db := config.DB

	// **🔹 Fetch Post with Associated Media & Categories**
	row := db.Table("posts").
		Select(`
			posts.id, posts.title, posts.content, posts.created_at, posts.updated_at,
			GROUP_CONCAT(DISTINCT categories.id ORDER BY categories.id ASC SEPARATOR ',') AS category_ids,
			GROUP_CONCAT(DISTINCT categories.name ORDER BY categories.id ASC SEPARATOR ',') AS category_names,
			GROUP_CONCAT(DISTINCT media.id ORDER BY media.id ASC SEPARATOR ',') AS media_ids,
			GROUP_CONCAT(DISTINCT media.path ORDER BY media.id ASC SEPARATOR ',') AS media_paths,
			GROUP_CONCAT(DISTINCT media.type ORDER BY media.id ASC SEPARATOR ',') AS media_types,
			GROUP_CONCAT(DISTINCT media.status ORDER BY media.id ASC SEPARATOR ',') AS media_statuses
		`).
		Joins(`
			LEFT JOIN post_categories ON post_categories.post_id = posts.id
			LEFT JOIN categories ON categories.id = post_categories.category_id
			LEFT JOIN media ON media.post_id = posts.id
		`).
		Where("posts.id = ?", postID).
		Group("posts.id").
		Row()

	// **🔹 Process Query Results**
	var (
		id                                              uint
		title                                           sql.NullString
		content                                         sql.NullString
		createdAt, updatedAt                            time.Time
		categoryIDs, categoryNames                      sql.NullString
		mediaIDs, mediaPaths, mediaTypes, mediaStatuses sql.NullString
	)

	// **🔹 Scan Row into Variables**
	err := row.Scan(&id, &title, &content, &createdAt, &updatedAt, &categoryIDs, &categoryNames, &mediaIDs, &mediaPaths, &mediaTypes, &mediaStatuses)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}

	// **🔹 Parse Categories**
	var categories []map[string]interface{}
	if categoryIDs.Valid && categoryNames.Valid {
		idList := strings.Split(categoryIDs.String, ",")
		nameList := strings.Split(categoryNames.String, ",")
		for i := range idList {
			categories = append(categories, map[string]interface{}{
				"id":   utils.ConvertStringToInt64(idList[i]),
				"name": nameList[i],
			})
		}
	}

	// **🔹 Parse Media (Separate Images & Videos)**
	var images []map[string]interface{}
	var videos []map[string]interface{}
	if mediaIDs.Valid && mediaPaths.Valid && mediaTypes.Valid && mediaStatuses.Valid {
		idList := strings.Split(mediaIDs.String, ",")
		pathList := strings.Split(mediaPaths.String, ",")
		typeList := strings.Split(mediaTypes.String, ",")
		statusList := strings.Split(mediaStatuses.String, ",")

		for i := range idList {
			status, _ := strconv.Atoi(statusList[i]) // Convert status to int
			mediaItem := map[string]interface{}{
				"id":     utils.ConvertStringToInt64(idList[i]),
				"path":   pathList[i],
				"type":   typeList[i],
				"status": status,
			}
			if strings.HasPrefix(typeList[i], "image/") {
				images = append(images, mediaItem)
			} else if strings.HasPrefix(typeList[i], "video/") {
				videos = append(videos, mediaItem)
			}
		}
	}

	// **🔹 Construct Final Response**
	postData := map[string]interface{}{
		"id":         id,
		"title":      title.String,
		"content":    content.String,
		"categories": categories,
		"images":     images,
		"videos":     videos,
		"created_at": createdAt,
		"updated_at": updatedAt,
	}

	return postData, nil
}
