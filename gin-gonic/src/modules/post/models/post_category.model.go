package models

import (
	"gorm.io/gorm"
	"time"
)

type PostCategory struct {
	ID         uint `gorm:"primary_key;AUTO_INCREMENT"`
	CategoryID uint `gorm:"not null"`
	PostID     uint `gorm:"not null"`

	// References
	Category Category `gorm:"foreignKey:CategoryID;references:ID"`
	Post     Post     `gorm:"foreignKey:PostID;references:ID"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func MigratePostCategory(db *gorm.DB) error {
	return db.AutoMigrate(&PostCategory{})
}
