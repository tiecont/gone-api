package models

import (
	"fmt"
	Roles "gone-be/modules/role/models"
	"gorm.io/gorm"
	"time"
)

// User represents the user model
type User struct {
	ID          uint       `gorm:"primaryKey;autoIncrement"`
	Username    string     `gorm:"size:255;not null;"`
	Email       string     `gorm:"size:255;not null;unique"`
	PhoneNumber int        `gorm:"size:11;default:null;unique"`
	GoogleID    string     `gorm:"size:255;default:null;unique"`
	TwitterID   string     `gorm:"size:255;default:null;unique"`
	Password    string     `gorm:"size:255;default:null"`
	Status      bool       `gorm:"default:true"`
	RoleID      uint       `gorm:"not null"`
	Role        Roles.Role `gorm:"foreignKey:RoleID;references:ID"` // Foreign key to Role
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
}

// MigrateUser migrates the User model to create the table in the database
func MigrateUser(db *gorm.DB) error {
	// Auto-migrate the User model
	if err := db.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("failed to auto-migrate User model: %w", err)
	}

	return nil
}
