package models

import (
	"time"

	"gorm.io/gorm"
)

// Role represents the role model.
type Role struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"unique;not null"` // Unique constraint ensures one-to-one relationship
	Name      string    `gorm:"type:char(30);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// MigrateRole migrates the Role model to create the table in the database.
func MigrateRole(db *gorm.DB) error {
	return db.AutoMigrate(&Role{})
}
