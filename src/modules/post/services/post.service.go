package services

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"gone-be/src/config"
	"gone-be/src/functions"
	"gone-be/src/modules/post/models"
	"gone-be/src/utils"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

func List(currentPage, itemsPerPage int, orderBy, sortBy, searchValue string) (map[string]interface{}, error) {
	// Validate sorting and calculate offset for pagination
	offsetData := utils.CalculateOffset(currentPage, itemsPerPage, sortBy, orderBy)
	// Fetch posts using the renewed SelectPosts function
	rows, err := functions.SelectPosts(
		searchValue,
		offsetData.OrderBy,
		offsetData.SortBy,
		offsetData.Offset,
		offsetData.ItemsPerPage,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Consolidate posts using a map keyed by post ID
	postMap := make(map[uint]map[string]interface{})
	var paginationResult map[string]interface{}

	for rows.Next() {
		// Define variables to scan row data
		var (
			postID, userID                                     uint
			content                                            sql.NullString
			createdAt, updatedAt                               time.Time
			categoryIDs, categoryNames                         sql.NullString
			mediaID                                            sql.NullInt64
			mediaStatus                                        sql.NullInt64
			mediaPath, mediaType                               sql.NullString
			username, profilePic                               sql.NullString
			commentsCount, likesCount, sharesCount, totalCount int
		)

		// Scan row values into the variables
		if err := rows.Scan(
			&postID, &content, &createdAt, &updatedAt,
			&userID, &username, &profilePic,
			&categoryIDs, &categoryNames,
			&mediaID, &mediaPath, &mediaType, &mediaStatus,
			&commentsCount, &likesCount, &sharesCount, &totalCount,
		); err != nil {
			return nil, err
		}

		// Check if the post is already in our map
		post, exists := postMap[postID]
		if !exists {
			// Parse category info into a slice if available
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

			// Initialize the post data
			post = map[string]interface{}{
				"id":             postID,
				"content":        content.String,
				"created_at":     createdAt,
				"updated_at":     updatedAt,
				"categories":     categories,
				"images":         []map[string]interface{}{},
				"videos":         []map[string]interface{}{},
				"user":           gin.H{"id": userID, "name": username.String, "profile_pic": profilePic.String},
				"comments_count": commentsCount,
				"likes_count":    likesCount,
				"shares_count":   sharesCount,
			}
			postMap[postID] = post
		}

		// Append media information to the appropriate slice
		if mediaID.Valid && mediaPath.Valid && mediaType.Valid && mediaStatus.Valid {
			mediaItem := map[string]interface{}{
				"id":     mediaID.Int64,
				"path":   mediaPath.String,
				"type":   mediaType.String,
				"status": int(mediaStatus.Int64),
			}
			if strings.HasPrefix(mediaType.String, "image/") {
				post["images"] = append(post["images"].([]map[string]interface{}), mediaItem)
			} else if strings.HasPrefix(mediaType.String, "video/") {
				post["videos"] = append(post["videos"].([]map[string]interface{}), mediaItem)
			}
		}

		// Use totalCount from any row (assumed same for all) to generate pagination metadata
		paginated, err := utils.Paginate(int64(totalCount), currentPage, itemsPerPage)
		if err != nil {
			return nil, err
		}
		paginationResult = paginated
	}

	// Convert the consolidated posts map to a slice for response
	items := make([]map[string]interface{}, 0, len(postMap))
	for _, post := range postMap {
		items = append(items, post)
	}

	return map[string]interface{}{
		"items":      items,
		"pagination": paginationResult,
	}, nil
}

// GetDetails is the function to get information details for a post with given postID
func GetDetails(postID string) (map[string]interface{}, error) {
	db := config.DB

	// **🔹 Query for Post with Associated Media & Categories**
	rows, err := db.Table("posts").
		Select(`
			posts.id, posts.content, posts.created_at, posts.updated_at,
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
		err := rows.Scan(&id, &content, &createdAt, &updatedAt, &categoryIDs, &categoryNames, &mediaID, &mediaPath, &mediaType, &mediaStatus)
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
		"content":    content.String,
		"categories": categories,
		"images":     images,
		"videos":     videos,
		"created_at": createdAt,
		"updated_at": updatedAt,
	}

	return postData, nil
}

type PostInfo struct {
	UserID      uint
	PostID      int64
	CategoryIDs []string
	Media       []*multipart.FileHeader
	Content     string
}

// CreatePost handles post creation along with media and categories
func CreatePost(content string, categoryIDs []string, files []*multipart.FileHeader, userID uint) (map[string]interface{}, *utils.ServiceError) {
	db := config.DB

	savedMediaResult, err := functions.SaveMediaToServer(files)
	if err != nil {
		return nil, &utils.ServiceError{StatusCode: err.StatusCode, Message: err.Message}
	}
	tx := db.Begin()
	if tx.Error != nil {
		return nil, &utils.ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to start transaction"}
	}

	// CreatePost
	postID, savePostErr := functions.CreatePost(content, userID)
	if savePostErr != nil {
		tx.Rollback()
		return nil, &utils.ServiceError{StatusCode: savePostErr.StatusCode, Message: savePostErr.Message}
	}
	// Save media to database
	saveMediaErr := functions.SaveMediaRecords(savedMediaResult, postID)
	if saveMediaErr != nil {
		tx.Rollback()
		return nil, &utils.ServiceError{StatusCode: saveMediaErr.StatusCode, Message: saveMediaErr.Message}
	}

	// Save Categories to Post
	saveCategoriesErr := functions.SaveCategoriesToPost(categoryIDs, postID)
	if saveCategoriesErr != nil {
		tx.Rollback()
		return nil, &utils.ServiceError{StatusCode: saveCategoriesErr.StatusCode, Message: saveCategoriesErr.Message}
	}
	tx.Commit()

	return map[string]interface{}{
		"id":         postID,
		"content":    content,
		"categories": categoryIDs,
	}, nil
}

// EditPostByUserID is the function to edits a post and updates its media and categories
func EditPostByUserID(postInfo PostInfo) *utils.ServiceError {
	db := config.DB
	tx := db.Begin()
	if tx.Error != nil {
		return &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to start transaction",
		}
	}
	// Check Post Exits
	var checkPostErr = functions.CheckPostExits(utils.ConvertInt64ToUint(postInfo.PostID))
	if checkPostErr != nil {
		tx.Rollback()
		return &utils.ServiceError{
			StatusCode: checkPostErr.StatusCode,
			Message:    checkPostErr.Message,
		}
	}

	// Delete existing media
	var existingMedia []models.Media
	if err := tx.Where("post_id = ?", postInfo.PostID).Find(&existingMedia).Error; err != nil {
		tx.Rollback()
		return &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to retrieve existing media",
		}
	}

	for _, media := range existingMedia {
		if err := functions.DeleteUploadedFile(media.Path); err != nil {
			tx.Rollback()
			return &utils.ServiceError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to delete old media",
			}
		}
	}
	// Delete existing media records
	deleteMediaErr := functions.DeleteMediaRecords(utils.ConvertInt64ToUint(postInfo.PostID))
	if deleteMediaErr != nil {
		tx.Rollback()
		return &utils.ServiceError{
			StatusCode: deleteMediaErr.StatusCode,
			Message:    deleteMediaErr.Message,
		}
	}
	// Delete existing category associations
	deleteCategoriesErr := functions.DeleteCategoryRecords(utils.ConvertInt64ToUint(postInfo.PostID))
	if deleteCategoriesErr != nil {
		tx.Rollback()
		return &utils.ServiceError{
			StatusCode: deleteCategoriesErr.StatusCode,
			Message:    deleteCategoriesErr.Message,
		}
	}

	// Update post content
	updatePostErr := functions.UpdatePostContent(postInfo.Content, utils.ConvertInt64ToUint(postInfo.PostID))
	if updatePostErr != nil {
		tx.Rollback()
		return &utils.ServiceError{
			StatusCode: updatePostErr.StatusCode,
			Message:    updatePostErr.Message,
		}
	}
	// New Media
	savedMediaResult, err := functions.SaveMediaToServer(postInfo.Media)
	if err != nil {
		tx.Rollback()
		return &utils.ServiceError{
			StatusCode: err.StatusCode, Message: err.Message,
		}
	}
	saveMediaRecordErr := functions.SaveMediaRecords(savedMediaResult, utils.ConvertInt64ToUint(postInfo.PostID))
	if saveMediaRecordErr != nil {
		tx.Rollback()
		return &utils.ServiceError{
			StatusCode: saveMediaRecordErr.StatusCode,
			Message:    saveMediaRecordErr.Message,
		}
	}

	// New Categories
	savedCategoriesErr := functions.SaveCategoriesToPost(postInfo.CategoryIDs, utils.ConvertInt64ToUint(postInfo.PostID))
	if savedCategoriesErr != nil {
		tx.Rollback()
		return &utils.ServiceError{
			StatusCode: savedCategoriesErr.StatusCode,
			Message:    savedCategoriesErr.Message,
		}
	}
	tx.Commit()

	return nil
}
