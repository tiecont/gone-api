package models

import (
	Users "gone-be/src/modules/user/models"
	"gorm.io/gorm"
	"time"
)

type CommentLike struct {
	ID        uint `gorm:"primary_key;AUTO_INCREMENT"`
	CommentID uint `gorm:"not null"`
	UserID    uint `gorm:"not null"`

	// References
	Comment Comment    `gorm:"foreignKey:CommentID;references:ID"`
	User    Users.User `gorm:"foreignKey:UserID;references:ID"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func MigrateCommentLike(db *gorm.DB) error {
	return db.AutoMigrate(&CommentLike{})
}
