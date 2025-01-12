package models

import (
	Users "gone-be/modules/user/models"
	"gorm.io/gorm"
	"time"
)

type Profile struct {
	ID         uint       `gorm:"primaryKey;autoIncrement"`
	UserID     uint       `gorm:"not null;unique"`                 // Foreign key referencing users
	User       Users.User `gorm:"foreignKey:UserID;references:ID"` // One-to-one relationship
	ProfilePic string     `gorm:"type:text"`                       // Profile picture URL
	CoverPhoto string     `gorm:"type:char(255)"`                  // Cover photo URL
	Sex        string     `gorm:"type:char(10)"`                   // Gender
	Birthday   time.Time  `gorm:"type:date"`                       // Date of birth
	Location   string     `gorm:"type:char(255)"`                  // Location
	Bio        string     `gorm:"type:text"`                       // Bio description
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime"`
}

// MigrateProfile migrates the Profile model to create the table in the database.
func MigrateProfile(db *gorm.DB) error {
	return db.AutoMigrate(&Profile{})
}
