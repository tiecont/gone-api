package models

import (
	Roles "gone-be/src/modules/role/models"
	"gorm.io/gorm"
	"time"
)

type Permission struct {
	ID        uint       `gorm:"primaryKey;AUTO_INCREMENT"`
	Name      string     `gorm:"type:varchar(255);not null"`
	RoleID    uint       `gorm:"not null"`
	Role      Roles.Role `gorm:"foreignKey:RoleID;references:ID"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

func MigratePermissions(db *gorm.DB) error {
	return db.AutoMigrate(&Permission{})
}
