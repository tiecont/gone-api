package services

import (
	"github.com/gin-gonic/gin"
	"gone-be/src/config"
	"gone-be/src/modules/permission/models"
)

// CheckPermission verifies if a role has a specific permission
func CheckPermission(roleID uint, permissionName string) bool {
	db := config.DB

	var exists bool
	err := db.Model(&models.Permission{}).
		Select("1").
		Where("role_id = ? AND name = ?", roleID, permissionName).
		Limit(1).
		Scan(&exists).Error

	if err != nil || !exists {
		return false
	}

	return exists
}

func CreatePermission(roleID uint, permissionName string) bool {
	db := config.DB
	exists := CheckPermission(roleID, permissionName)
	if exists {
		return false
	}
	var permission models.Permission
	permission.RoleID = roleID
	permission.Name = permissionName

	if err := db.Create(&permission).Error; err != nil {
		return false
	}
	return true
}

func ListAllPermission(roleID uint) ([]map[string]interface{}, error) {
	db := config.DB
	var permissions []models.Permission

	// Fetch permissions from the database
	if err := db.Where("role_id = ?", roleID).Select("id, name").Find(&permissions).Error; err != nil {
		return nil, err
	}

	// Convert to a slice of maps
	var result []map[string]interface{}
	for _, permission := range permissions {
		result = append(result, gin.H{
			"id":   permission.ID,
			"name": permission.Name,
		})
	}

	return result, nil
}
