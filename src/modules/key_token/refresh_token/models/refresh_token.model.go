package models

import (
	Users "gone-be/src/modules/user/models"
	"time"

	"gorm.io/gorm"
)

// RefreshToken represents the refresh token model.
type RefreshToken struct {
	ID        uint       `gorm:"primaryKey;autoIncrement"`
	UserID    uint       `gorm:"not null"`
	User      Users.User `gorm:"foreignKey:UserID;references:ID"`
	Token     string     `gorm:"type:varchar(256);not null"`
	Status    bool       `gorm:"default:true"` // true = active, false = revoked
	ExpiresAt time.Time  `gorm:"not null"`     // Token expiration time
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

// BeforeCreate sets the default expiration time for the refresh token.
func (r *RefreshToken) BeforeCreate(tx *gorm.DB) (err error) {
	r.ExpiresAt = time.Now().Add(30 * 24 * time.Hour) // 30 days from now
	return
}

// MigrateRefreshTokens migrates the RefreshToken model.
func MigrateRefreshTokens(db *gorm.DB) error {
	return db.AutoMigrate(&RefreshToken{})
}
