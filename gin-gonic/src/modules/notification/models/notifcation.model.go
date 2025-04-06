package models

import (
	Users "gone-be/src/modules/user/models"
	"gorm.io/gorm"
	"time"
)

type Notification struct {
	ID         uint       `gorm:"primary_key;AUTO_INCREMENT"`
	SenderID   uint       `gorm:"not null"`
	ReceiverID uint       `gorm:"not null"`
	Sender     Users.User `gorm:"foreignKey:SenderID;references:ID"`
	Receiver   Users.User `gorm:"foreignKey:ReceiverID;references:ID"`
	Message    string     `gorm:"type:text;default:null"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime"`
}

func MigrateNotification(db *gorm.DB) error {
	return db.AutoMigrate(&Notification{})
}
