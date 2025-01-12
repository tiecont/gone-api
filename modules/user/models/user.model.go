package models

import (
	Roles "gone-be/modules/role/models"
	"gorm.io/gorm"
	"time"
)

// User represents the user model.
type User struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	Username  string     `gorm:"size:255;not null;unique"`
	Email     string     `gorm:"size:255;not null;unique"`
	Password  string     `gorm:"size:255;not null"`
	Status    bool       `gorm:"default:true"`
	RoleID    uint       `gorm:"not null"`
	Role      Roles.Role `gorm:"foreignKey:RoleID;references:ID"` // Foreign key to Role
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

// MigrateUser migrates the User model to create the table in the database
func MigrateUser(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
