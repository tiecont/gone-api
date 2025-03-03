package models

import (
	Permissions "gone-be/src/modules/permission/models"
	"gorm.io/gorm"
	"time"
)

type RolePermission struct {
	ID           uint                   `gorm:"primary_key;AUTO_INCREMENT"`
	RoleID       uint                   `gorm:"not null"`
	PermissionID uint                   `gorm:"not null"`
	Role         Role                   `gorm:"foreignKey:RoleID;references:ID"`
	Permission   Permissions.Permission `gorm:"foreignKey:PermissionID;references:ID"`
	CreatedAt    time.Time              `gorm:"autoCreateTime"`
	UpdatedAt    time.Time              `gorm:"autoUpdateTime"`
}

// MigrateRolePermissions migrates the RolePermissions model to create the table in the database.
func MigrateRolePermissions(db *gorm.DB) error {
	return db.AutoMigrate(&RolePermission{})
}
