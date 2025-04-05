package services

import (
	"gone-be/src/config"
	"gone-be/src/functions"
	"gone-be/src/modules/post/models"
	"gone-be/src/utils"
	"gorm.io/gorm"
	"net/http"
)

// Like is a service function to process the like action.
func Like(userID, postID uint) (int64, *utils.ServiceError) {
	db := config.DB
	var counts int64

	// 1. Check if the user already liked the post
	liked, err := functions.CheckIsLiked(userID, postID)
	if err != nil {
		return counts, &utils.ServiceError{Message: err.Message, StatusCode: err.StatusCode}
	}

	if liked {
		return counts, &utils.ServiceError{
			Message:    "You have already liked this post",
			StatusCode: http.StatusBadRequest,
		}
	}

	// 2. Create a new like for the post in a single query
	newLike := models.PostLike{
		UserID: userID,
		PostID: postID,
	}

	// Insert the like, while also counting the total likes in one query.
	// Using db.Transaction to ensure atomicity of the operations.
	if err := db.Transaction(func(tx *gorm.DB) error {
		// Insert the like
		if err := tx.Create(&newLike).Error; err != nil {
			return err
		}

		// Count the total likes after inserting the new like
		if err := tx.Model(&models.PostLike{}).
			Where("post_id = ?", postID).
			Count(&counts).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return counts, &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to like the post",
		}
	}

	return counts, nil
}

// DisLike is a service function to delete a like from the PostLike table and update the like count.
func DisLike(userID, postID uint) (int64, *utils.ServiceError) {
	db := config.DB
	var counts int64

	// 1. Check if the user has already liked the post
	liked, err := functions.CheckIsLiked(userID, postID)
	if err != nil {
		return counts, &utils.ServiceError{Message: err.Message, StatusCode: err.StatusCode}
	}

	if !liked {
		return counts, &utils.ServiceError{
			Message:    "You have already disliked this post",
			StatusCode: http.StatusBadRequest,
		}
	}

	// 2. Delete the like record for the post in a single query
	if err := db.Transaction(func(tx *gorm.DB) error {
		// Delete the like
		if err := tx.Where("post_id = ? AND user_id = ?", postID, userID).
			Delete(&models.PostLike{}).Error; err != nil {
			return err
		}

		// Count the total likes after deleting the like
		if err := tx.Model(&models.PostLike{}).
			Where("post_id = ?", postID).
			Count(&counts).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return counts, &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to delete the liked post",
		}
	}

	return counts, nil
}
