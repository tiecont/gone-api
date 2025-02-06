package models

import (
	"gorm.io/gorm"
	"time"
)

type Category struct {
	ID        uint      `gorm:"primary_key;AUTO_INCREMENT"`
	Name      string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func MigrateCategory(db *gorm.DB) error {
	return db.AutoMigrate(&Category{})
}
