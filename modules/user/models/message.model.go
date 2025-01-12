package models

import (
	"gorm.io/gorm"
	"time"
)

type Message struct {
	ID         uint `gorm:"primary_key;AUTO_INCREMENT"`
	SenderID   uint `gorm:"not null"`
	ReceiverID uint `gorm:"not null"`

	// References
	Sender   User `gorm:"foreignKey:SenderID"`
	Receiver User `gorm:"foreignKey:ReceiverID"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func MigrateMessage(db *gorm.DB) error {
	return db.AutoMigrate(&Message{})
}
