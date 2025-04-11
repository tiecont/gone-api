package services

import (
	"database/sql"
	"gone-be/src/config"
	"gone-be/src/functions"
	"gone-be/src/modules/notification/models"
	"gone-be/src/services"
	"gone-be/src/utils"
	"net/http"
	"time"
)

func NotificationHandler(senderID, receiverID uint, message, noti_type string) *utils.ServiceError {
	db := config.DB

	// Prepare for new notification record
	newNoti := models.Notification{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Message:    message,
		NotiType:   noti_type,
	}

	// Inserting record
	if err := db.Create(&newNoti).Error; err != nil {
		return &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}
	// Prepare message
	content := services.WebSocketMessage{
		Type:    newNoti.NotiType,
		Message: newNoti.Message,
	}
	// Send Notification on Socket
	err := services.SendMessageToUser(receiverID, content)
	if err != nil {
		return nil
	}
	return nil
}

func GetNotificationsByUserID(userID uint, currentPage, itemsPerPage int, orderBy, sortBy, searchValue string, isSeen bool, all bool) (map[string]interface{}, *utils.ServiceError) {
	// Prepare Pagination
	offsetData := utils.CalculateOffset(currentPage, itemsPerPage, sortBy, orderBy)

	// Get Rows
	rows, err := functions.ListNotifications(searchValue, offsetData.OrderBy, offsetData.SortBy, offsetData.Offset, itemsPerPage, isSeen, userID, all)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notificationMap := make(map[uint]map[string]interface{})
	var paginationResult map[string]interface{}

	for rows.Next() {
		// Declare variables for scanning
		var (
			notificationID, senderID, receiverID uint
			message                              sql.NullString
			createdAt, updatedAt                 time.Time
			isSeen                               sql.NullBool
			notiType                             sql.NullString
			totalCount                           int
		)
		if err := rows.Scan(
			&notificationID, &senderID, &receiverID,
			&message, &createdAt, &updatedAt, &isSeen, &notiType,
			&totalCount); err != nil {
			return nil, &utils.ServiceError{
				StatusCode: http.StatusInternalServerError,
				Message:    err.Error(),
			}
		}

		// If post doesn't exist in map, initialize it
		notifications, exists := notificationMap[notificationID]
		if !exists {
			notifications = map[string]interface{}{
				"id":         notificationID,
				"sender_id":  senderID,
				"message":    message.String,
				"read":       isSeen.Bool,
				"type":       notiType.String,
				"created_at": createdAt.String(),
				"updated_at": updatedAt.String(),
			}
			notificationMap[notificationID] = notifications
		}

		// Build pagination metadata once
		if paginationResult == nil {
			paginated, err := utils.Paginate(int64(totalCount), currentPage, itemsPerPage)
			if err != nil {
				return nil, &utils.ServiceError{
					StatusCode: http.StatusInternalServerError,
					Message:    err.Error(),
				}
			}
			paginationResult = paginated
		}
	}
	// Convert postMap to slice
	items := make([]map[string]interface{}, 0, len(notificationMap))
	for _, post := range notificationMap {
		items = append(items, post)
	}

	return map[string]interface{}{
		"items":      items,
		"pagination": paginationResult,
	}, nil
}

func UpdateIsSeen(notificationID, userID uint) *utils.ServiceError {
	db := config.DB

	if err := db.Model(&models.Notification{}).
		Where("id = ? AND receiver_id = ?", notificationID, userID).
		Update("is_seen", true).Error; err != nil {
		return &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to update is_seen status",
		}
	}

	return nil
}

func UpdateIsSeenForAllNotificationByUserID(userID uint) *utils.ServiceError {
	db := config.DB

	if err := db.Model(&models.Notification{}).
		Where("receiver_id = ? AND is_seen = false", userID).
		Update("is_seen", true).Error; err != nil {
		return &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to update is_seen status",
		}
	}

	return nil
}
