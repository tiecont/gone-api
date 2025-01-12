package models

import (
	Users "gone-be/modules/user/models"
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	ID     uint   `gorm:"primary_key;AUTO_INCREMENT"`
	PostID uint   `gorm:"not null"`
	UserID uint   `gorm:"not null"`
	Text   string `gorm:"type:text;not null"`
	Left   int    `gorm:"not null"`
	Right  int    `gorm:"not null"`

	// References
	Post Post       `gorm:"foreignKey:PostID;references:ID"`
	User Users.User `gorm:"foreignKey:UserID;references:ID"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func MigrateComment(db *gorm.DB) error {
	return db.AutoMigrate(&Comment{})
}
