package models

import (
	Users "gone-be/src/modules/user/models"
	"gorm.io/gorm"
	"time"
)

type PostLike struct {
	ID     uint `gorm:"primary_key;AUTO_INCREMENT"`
	PostID uint `gorm:"not null"`
	UserID uint `gorm:"not null"`

	// References
	Post Post       `gorm:"foreignKey:PostID;references:ID"`
	User Users.User `gorm:"foreignKey:UserID;references:ID"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func MigratePostLike(db *gorm.DB) error {
	return db.AutoMigrate(&PostLike{})
}
