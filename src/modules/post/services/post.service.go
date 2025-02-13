package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gone-be/src/config"
	"gone-be/src/modules/post/models"
	"gone-be/src/utils"
	"io"
	"mime/multipart"
	"net/http"
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
	"image/webp": true,
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

func CreatePost(title string, content string, categoryIDs []string, files []*multipart.FileHeader, userID uint) (map[string]interface{}, *utils.ServiceError) {
	db := config.DB

	var savedMedia []models.Media

	for _, file := range files {
		// Validate file type
		fileType := file.Header.Get("Content-Type")
		if !allowedMediaTypes[fileType] {
			return nil, &utils.ServiceError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid file format. Allowed: JPEG, PNG, GIF, MP4, AVI, MOV",
			}
		}

		// Generate unique file name
		uniqueFileName := time.Now().Format("20060102150405") + "_" + file.Filename
		mediaPath := filepath.Join("uploads", uniqueFileName)

		// Save file to disk
		if err := SaveUploadedFile(file, mediaPath); err != nil {
			return nil, &utils.ServiceError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to save file",
			}
		}

		// Append to saved media list
		savedMedia = append(savedMedia, models.Media{
			Path: mediaPath,
			Type: fileType,
		})
	}

	// Start database transaction
	tx := db.Begin()
	if tx.Error != nil {
		return nil, &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to start transaction",
		}
	}

	// Create post
	post := models.Post{
		UserID:  userID,
		Title:   title,
		Content: content,
	}
	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		return nil, &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to create post",
		}
	}

	// Save media records (Images & Videos)
	for _, media := range savedMedia {
		media.PostID = post.ID
		if err := tx.Create(&media).Error; err != nil {
			tx.Rollback()
			return nil, &utils.ServiceError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to save media",
			}
		}
	}

	// Convert categoryIDs to integers & create relationships
	var postCategories []models.PostCategory
	for _, categoryID := range categoryIDs {
		id, err := strconv.Atoi(categoryID)
		if err != nil {
			tx.Rollback()
			return nil, &utils.ServiceError{
				StatusCode: http.StatusBadRequest,
				Message:    "Invalid category ID",
			}
		}
		postCategories = append(postCategories, models.PostCategory{
			PostID:     post.ID,
			CategoryID: uint(id),
		})
	}

	if len(postCategories) > 0 {
		if err := tx.Create(&postCategories).Error; err != nil {
			tx.Rollback()
			return nil, &utils.ServiceError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to associate categories",
			}
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
			users.id AS user_id, users.username AS username, profiles.profile_pic,
			GROUP_CONCAT(DISTINCT categories.id ORDER BY categories.id ASC SEPARATOR ',') AS category_ids,
			GROUP_CONCAT(DISTINCT categories.name ORDER BY categories.id ASC SEPARATOR ',') AS category_names,
			media.id, media.path, media.type, media.status,
			COUNT(DISTINCT comments.id) AS comment_count,
			COUNT(DISTINCT post_likes.id) AS likes_count,
			COUNT(DISTINCT post_shares.id) AS shares_count
		`).
		Joins(`
			LEFT JOIN post_categories ON post_categories.post_id = posts.id
			LEFT JOIN categories ON categories.id = post_categories.category_id
			LEFT JOIN media ON media.post_id = posts.id
			LEFT JOIN users ON users.id = posts.user_id
			LEFT JOIN profiles ON profiles.user_id = users.id
			LEFT JOIN comments ON comments.post_id = posts.id
			LEFT JOIN post_likes ON post_likes.post_id = posts.id
			LEFT JOIN post_shares ON post_shares.post_id = posts.id
		`).
		Where("LOWER(posts.content) LIKE LOWER(?)", "%"+searchValue+"%").
		Group("posts.id, users.id, users.username, profiles.profile_pic, media.id").
		Order(fmt.Sprintf("posts.%s %s", orderBy, sortBy)).
		Offset(offset).
		Limit(itemsPerPage).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// **🔹 Process Query Results**
	postMap := make(map[uint]map[string]interface{})
	var totalCount int64

	for rows.Next() {
		var postID, userID uint
		var title, content sql.NullString
		var createdAt, updatedAt time.Time
		var categoryIDs, categoryNames sql.NullString
		var mediaID, mediaStatus sql.NullInt64
		var mediaPath, mediaType sql.NullString
		var username, profilePic sql.NullString
		var commentsCount, likesCount, sharesCount int

		// **🔹 Scan Row**
		if err := rows.Scan(
			&postID, &title, &content, &createdAt, &updatedAt,
			&userID, &username, &profilePic,
			&categoryIDs, &categoryNames,
			&mediaID, &mediaPath, &mediaType, &mediaStatus,
			&commentsCount, &likesCount, &sharesCount,
		); err != nil {
			return nil, err
		}

		// **🔹 Check if Post Already Exists in Map**
		post, exists := postMap[postID]
		if !exists {
			// **🔹 Parse Categories (Only Once Per Post)**
			var categories []map[string]interface{}
			if categoryIDs.Valid && categoryNames.Valid {
				idList := strings.Split(categoryIDs.String, ",")
				nameList := strings.Split(categoryNames.String, ",")
				for i := range idList {
					if i < len(nameList) {
						categories = append(categories, map[string]interface{}{
							"id":   utils.ConvertStringToInt64(idList[i]),
							"name": nameList[i],
						})
					}
				}
			}

			// **🔹 Initialize Post Data**
			post = map[string]interface{}{
				"id":         postID,
				"title":      title.String,
				"content":    content.String,
				"categories": categories,
				"created_at": createdAt,
				"updated_at": updatedAt,
				"images":     []map[string]interface{}{},
				"videos":     []map[string]interface{}{},
				"user": gin.H{
					"id":          userID,
					"name":        username.String,
					"profile_pic": profilePic.String,
				},
				"comments_count": commentsCount,
				"likes_count":    likesCount,
				"shares_count":   sharesCount,
			}

			postMap[postID] = post
			totalCount++
		}

		// **🔹 Process Media (Append to Existing Post)**
		if mediaID.Valid && mediaPath.Valid && mediaType.Valid && mediaStatus.Valid {
			status := int(mediaStatus.Int64)
			mediaItem := map[string]interface{}{
				"id":     mediaID.Int64,
				"path":   mediaPath.String,
				"type":   mediaType.String,
				"status": status,
			}

			if strings.HasPrefix(mediaType.String, "image/") {
				post["images"] = append(post["images"].([]map[string]interface{}), mediaItem)
			} else if strings.HasPrefix(mediaType.String, "video/") {
				post["videos"] = append(post["videos"].([]map[string]interface{}), mediaItem)
			}
		}
	}

	// **🔹 Convert Map to Slice**
	items := make([]map[string]interface{}, 0, len(postMap))
	for _, post := range postMap {
		items = append(items, post)
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

	// **🔹 Query for Post with Associated Media & Categories**
	rows, err := db.Table("posts").
		Select(`
			posts.id, posts.title, posts.content, posts.created_at, posts.updated_at,
			GROUP_CONCAT(DISTINCT categories.id ORDER BY categories.id ASC SEPARATOR ',') AS category_ids,
			GROUP_CONCAT(DISTINCT categories.name ORDER BY categories.id ASC SEPARATOR ',') AS category_names,
			media.id, media.path, media.type, media.status
		`).
		Joins(`
			LEFT JOIN post_categories ON post_categories.post_id = posts.id
			LEFT JOIN categories ON categories.id = post_categories.category_id
			LEFT JOIN media ON media.post_id = posts.id
		`).
		Where("posts.id = ?", postID).
		Group("posts.id, media.id").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// **🔹 Initialize Storage for Data**
	var (
		id                         uint
		title                      sql.NullString
		content                    sql.NullString
		createdAt                  time.Time
		updatedAt                  time.Time
		categoryIDs, categoryNames sql.NullString
	)

	var categories []map[string]interface{}
	var images []map[string]interface{}
	var videos []map[string]interface{}

	// **🔹 Iterate Over Rows to Collect Data**
	for rows.Next() {
		var (
			mediaID, mediaStatus int64
			mediaPath, mediaType sql.NullString
		)

		// **Scan Data into Variables**
		err := rows.Scan(&id, &title, &content, &createdAt, &updatedAt, &categoryIDs, &categoryNames, &mediaID, &mediaPath, &mediaType, &mediaStatus)
		if err != nil {
			return nil, err
		}

		// **Process Categories Only Once**
		if len(categories) == 0 && categoryIDs.Valid && categoryNames.Valid {
			idList := strings.Split(categoryIDs.String, ",")
			nameList := strings.Split(categoryNames.String, ",")
			for i := range idList {
				categories = append(categories, map[string]interface{}{
					"id":   utils.ConvertStringToInt64(idList[i]),
					"name": nameList[i],
				})
			}
		}

		// **Process Media**
		if mediaPath.Valid && mediaType.Valid {
			mediaItem := map[string]interface{}{
				"id":     mediaID,
				"path":   mediaPath.String,
				"type":   mediaType.String,
				"status": mediaStatus,
			}

			if strings.HasPrefix(mediaType.String, "image/") {
				images = append(images, mediaItem)
			} else if strings.HasPrefix(mediaType.String, "video/") {
				videos = append(videos, mediaItem)
			}
		}
	}

	// **🔹 If No Rows Found, Return Error**
	if id == 0 {
		return nil, errors.New("post not found")
	}

	// **🔹 Construct Final Response**
	postData := map[string]interface{}{
		"id":         id,
		"title":      title.String,
		"content":    content.String,
		"categories": categories,
		"images":     images,
		"videos":     videos,
		"created_at": createdAt.Format(time.RFC3339),
		"updated_at": updatedAt.Format(time.RFC3339),
	}

	return postData, nil
}
