package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents the user model.
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Username  string    `gorm:"size:255;not null;unique"`
	Email     string    `gorm:"size:255;not null;unique"`
	Password  string    `gorm:"size:255;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// MigrateUser migrates the User model to create the table in the database.
func MigrateUser(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
