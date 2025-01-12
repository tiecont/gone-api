package models

import (
	Users "gone-be/modules/user/models"
	"time"

	"gorm.io/gorm"
)

// AccessToken represents the access token model.
type AccessToken struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	UserID    uint       `gorm:"not null"`
	User      Users.User `gorm:"foreignKey:UserID;references:ID"`
	Token     string     `gorm:"type:varchar(256);not null"`
	Status    bool       `gorm:"default:true"` // true = active, false = revoked
	ExpiresAt time.Time  `gorm:"not null"`     // Token expiration time
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

// BeforeCreate sets the default expiration time for the access token.
func (a *AccessToken) BeforeCreate(tx *gorm.DB) (err error) {
	a.ExpiresAt = time.Now().Add(7 * 24 * time.Hour) // 7 days from now
	return
}

// MigrateAccessTokens migrates the AccessToken model.
func MigrateAccessTokens(db *gorm.DB) error {
	return db.AutoMigrate(&AccessToken{})
}
