package models

import (
	Users "gone-be/src/modules/user/models"
	"gorm.io/gorm"
	"time"
)

type Post struct {
	ID        uint       `gorm:"primaryKey;AUTO_INCREMENT"`
	UserID    uint       `gorm:"not null"`
	User      Users.User `gorm:"foreignKey:UserID;references:ID"`
	Content   string     `gorm:"type:text;default:null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

func MigratePost(db *gorm.DB) error {
	return db.AutoMigrate(&Post{})
}
