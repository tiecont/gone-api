package models

import (
	"time"

	"gorm.io/gorm"
)

// Friend represents the friends table, establishing relationships between users.
type Friend struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	UserID      uint      `gorm:"not null"` // The user who initiated the friend request
	FriendTo    uint      `gorm:"not null"` // The user who is the friend
	RequestedOn time.Time `gorm:"not null"`
	AcceptedOn  time.Time `gorm:"default:null"`
	Description string    `gorm:"type:text;default:null"`
	Status      bool      `gorm:"default:false"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`

	// Relationships
	User   User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`   // Relation to the User who initiated the request
	Friend User `gorm:"foreignKey:FriendTo;constraint:OnDelete:CASCADE;"` // Relation to the User who is the friend
}

// MigrateFriends migrates the Friend model to create the friends table.
func MigrateFriends(db *gorm.DB) error {
	return db.AutoMigrate(&Friend{})
}
