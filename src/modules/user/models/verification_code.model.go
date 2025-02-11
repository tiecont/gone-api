package models

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type VerificationCode struct {
	ID         uint      `gorm:"primary_key;AUTO_INCREMENT"`
	Email      string    `gorm:"type:varchar(255);not null"`
	Code       string    `gorm:"type:varchar(255);not null"`
	ExpiresAt  time.Time `gorm:"type:datetime;not null"`
	InputCount int       `gorm:"type:int(5);not null"`
	CreatedAt  time.Time `gorm:"autoCreateTime;"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime;"`
}

func MigrateVerificationCode(db *gorm.DB) error {
	if err := db.AutoMigrate(&VerificationCode{}); err != nil {
		return fmt.Errorf("failed to auto-migrate User model: %w", err)
	}

	return nil
}
