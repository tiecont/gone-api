package models

import (
	Users "gone-be/modules/user/models"
	"gorm.io/gorm"
	"time"
)

type Post struct {
	ID        uint       `gorm:"primaryKey;AUTO_INCREMENT"`
	UserID    uint       `gorm:"not null"`
	User      Users.User `gorm:"foreignKey:UserID;references:ID"`
	title     string     `gorm:"size:255;default:null"`
	content   string     `gorm:"type:text;default:null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

func MigratePost(db *gorm.DB) error {
	return db.AutoMigrate(&Post{})
}
