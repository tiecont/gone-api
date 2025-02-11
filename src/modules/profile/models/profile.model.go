package models

import (
	Users "gone-be/src/modules/user/models"
	"gorm.io/gorm"
	"time"
)

type Profile struct {
	ID              uint       `gorm:"primaryKey;autoIncrement"`
	UserID          uint       `gorm:"not null;unique"`                 // Foreign key referencing users
	User            Users.User `gorm:"foreignKey:UserID;references:ID"` // One-to-one relationship
	ProfilePic      string     `gorm:"type:text;default:null"`          // Profile picture URL
	CoverPhoto      string     `gorm:"type:char(255);default:null"`     // Cover photo URL
	BackgroundColor string     `gorm:"type:char(255);default:#7b2cbf"`
	Sex             string     `gorm:"type:char(10);default:null"`  // Gender
	Birthday        *time.Time `gorm:"type:date"`                   // Date of birth
	Location        string     `gorm:"type:char(255);default:null"` // Location
	Bio             string     `gorm:"type:text;default:null"`      // Bio description
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"`
}

// MigrateProfile migrates the Profile model to create the table in the database.
func MigrateProfile(db *gorm.DB) error {
	return db.AutoMigrate(&Profile{})
}
