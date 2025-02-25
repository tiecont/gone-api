package functions

import (
	"database/sql"
	"fmt"
	"gone-be/src/config"
	"gone-be/src/modules/post/models"
	"gone-be/src/utils"
	"net/http"
	"strconv"
)

// CreatePost is the function will create post with userID and content
func CreatePost(content string, userID uint) (postID uint, errService *utils.ServiceError) {
	db := config.DB
	post := models.Post{UserID: userID, Content: content}
	if err := db.Create(&post).Error; err != nil {
		return 0, &utils.ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to create post"}
	}
	return post.ID, nil
}

// SaveCategoriesToPost is the function will save post categories to database
func SaveCategoriesToPost(categoryIDs []string, postID uint) *utils.ServiceError {
	db := config.DB
	var postCategories []models.PostCategory
	for _, categoryID := range categoryIDs {
		id, err := strconv.Atoi(categoryID)
		if err != nil {
			return &utils.ServiceError{StatusCode: http.StatusBadRequest, Message: "Invalid category ID"}
		}
		postCategories = append(postCategories, models.PostCategory{PostID: postID, CategoryID: uint(id)})
	}

	if len(postCategories) > 0 {
		if err := db.Create(&postCategories).Error; err != nil {
			return &utils.ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to associate categories"}
		}
	}
	return nil
}

// SaveMediaRecords is the function will save from media path into database
func SaveMediaRecords(savedMedia []models.Media, postID uint) *utils.ServiceError {
	db := config.DB
	for _, media := range savedMedia {
		media.PostID = postID
		if err := db.Create(&media).Error; err != nil {
			return &utils.ServiceError{StatusCode: http.StatusInternalServerError, Message: "Failed to save media"}
		}
	}
	return nil
}

// DeletePostRecord is the function to delete post records from database
func DeletePostRecord(postID uint) *utils.ServiceError {
	db := config.DB
	// Attempt to delete the post record with the given ID
	if err := db.Delete(&models.Post{}, postID).Error; err != nil {
		return &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to delete post",
		}
	}
	return nil
}

// DeleteMediaRecords is the function to delete media records following to the post
func DeleteMediaRecords(postID uint) *utils.ServiceError {
	db := config.DB

	if err := db.Where("post_id = ?", postID).Delete(&models.Media{}).Error; err != nil {
		return &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to remove old media records",
		}
	}
	return nil
}

// DeleteCategoryRecords is the function to delete categories following to the post
func DeleteCategoryRecords(postID uint) *utils.ServiceError {
	db := config.DB
	if err := db.Where("post_id = ?", postID).Delete(&models.PostCategory{}).Error; err != nil {
		return &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to remove old category associations",
		}
	}
	return nil
}

// UpdatePostContent is the function to update the content of post by given postID
func UpdatePostContent(content string, postID uint) *utils.ServiceError {
	db := config.DB
	if err := db.Model(&models.Post{}).Where("id = ?", postID).Updates(models.Post{
		Content: content,
	}).Error; err != nil {
		return &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to update post",
		}
	}
	return nil
}

// CheckPostExits is the function will check post was valid on database or not
func CheckPostExits(postID uint) *utils.ServiceError {
	db := config.DB
	// Check if Post Exists
	var exists bool
	if err := db.Model(&models.Post{}).
		Select("count(*) > 0").
		Where("id = ?", postID).
		Find(&exists).Error; err != nil {

		return &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to check post existence",
		}
	}

	// If the post does not exist, return an error
	if !exists {
		return &utils.ServiceError{
			StatusCode: http.StatusNotFound,
			Message:    "Post not found",
		}
	}
	return nil
}

// SelectPosts is the function will execute the sql queries with given parameters and return rows
func SelectPosts(searchValue, orderBy, sortBy string, offset, itemsPerPage int) (*sql.Rows, *utils.ServiceError) {
	rows, err := config.DB.Table("posts").
		Select(`
			posts.id, posts.content, posts.created_at, posts.updated_at,
			users.id AS user_id, users.username AS username, profiles.profile_pic,
			GROUP_CONCAT(DISTINCT categories.id ORDER BY categories.id ASC SEPARATOR ',') AS category_ids,
			GROUP_CONCAT(DISTINCT categories.name ORDER BY categories.id ASC SEPARATOR ',') AS category_names,
			media.id AS media_id, media.path, media.type, media.status,
			COUNT(DISTINCT comments.id) AS comment_count,
			COUNT(DISTINCT post_likes.id) AS likes_count,
			COUNT(DISTINCT post_shares.id) AS shares_count,
			COUNT(posts.id) OVER() AS total_count
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
		return nil, &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}
	return rows, nil
}
