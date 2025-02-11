package models

import (
	"gorm.io/gorm"
	"time"
)

type Media struct {
	ID        uint      `gorm:"primary_key;AUTO_INCREMENT"`
	PostID    uint      `gorm:"not null"`
	Post      Post      `gorm:"foreignKey:PostID;references:ID;constraint:OnDelete:CASCADE;"`
	Path      string    `gorm:"type:text;not null"`
	Type      string    `gorm:"size:50;not null"`
	Status    bool      `gorm:"default:true"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func MigrateMedia(db *gorm.DB) error {
	return db.AutoMigrate(&Media{})
}
