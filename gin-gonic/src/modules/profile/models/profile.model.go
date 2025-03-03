package models

import (
	Users "gone-be/src/modules/user/models"
	"gorm.io/gorm"
	"time"
)

type Profile struct {
	ID              uint       `gorm:"primaryKey;autoIncrement"`
	UserID          uint       `gorm:"not null;unique"`
	User            Users.User `gorm:"foreignKey:UserID;references:ID"`
	ProfilePic      string     `gorm:"type:text;default:null"`
	CoverPhoto      string     `gorm:"type:char(255);default:null"`
	BackgroundColor string     `gorm:"type:char(255);default:#7b2cbf"`
	Gender          string     `gorm:"type:char(10);default:null"`
	Birthday        *time.Time `gorm:"type:date"`
	Location        string     `gorm:"type:char(255);default:null"`
	Bio             string     `gorm:"type:text;default:null"`
	Interests       []string   `gorm:"type:json;default:null"`
	SocialLinks     []string   `gorm:"type:json;default:null"`

	// Many-to-Many Relationship (Followers & Followings)
	Followers  []Users.User `gorm:"many2many:follows;joinForeignKey:FollowerID;joinReferences:FollowingID"`
	Followings []Users.User `gorm:"many2many:follows;joinForeignKey:FollowingID;joinReferences:FollowerID"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// MigrateProfile migrates the Profile model
func MigrateProfile(db *gorm.DB) error {
	return db.AutoMigrate(&Profile{})
}
