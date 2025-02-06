package models

import (
	"gorm.io/gorm"
	"time"
)

type Permission struct {
	ID        uint      `gorm:"primaryKey;AUTO_INCREMENT"`
	Name      string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func MigratePermissions(db *gorm.DB) error {
	return db.AutoMigrate(&Permission{})
}
