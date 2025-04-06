package services

import (
	"gone-be/src/config"
	"gone-be/src/modules/notification/models"
	"gone-be/src/services"
	"gone-be/src/utils"
	"net/http"
)

func NotificationHandler(senderID, receiverID uint, message string) *utils.ServiceError {
	db := config.DB

	// Prepare for new notification record
	newNoti := models.Notification{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Message:    message,
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
		Type:    "notice",
		Message: newNoti.Message,
	}
	// Send Notification on Socket
	err := services.SendMessageToUser(receiverID, content)
	if err != nil {
		return nil
	}
	return nil
}
