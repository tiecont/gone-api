package services

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gone-be/src/config"
	"gone-be/src/modules/role/models"
	Users "gone-be/src/modules/user/models"
	"gorm.io/gorm"
)

// GetRoleByUserID retrieves the role of a user by their user ID
func GetRoleByUserID(userID uint) (*models.Role, error) {
	db := config.DB

	// Fetch user to get RoleID
	var user Users.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Fetch role based on RoleID
	var role models.Role
	if err := db.Where("id = ?", user.RoleID).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	return &role, nil
}

// CreateRoleByAdmin creates a new role with the given name
func CreateRoleByAdmin(roleName string) (*models.Role, error) {
	db := config.DB

	// Validate input
	if roleName == "" {
		return nil, errors.New("role name cannot be empty")
	}

	// Check if role already exists
	var existingRole models.Role
	if err := db.Where("name = ?", roleName).First(&existingRole).Error; err == nil {
		return nil, errors.New("role already exists")
	}

	// Create new role
	role := &models.Role{
		Name: roleName,
	}

	if err := db.Create(role).Error; err != nil {
		return nil, err
	}

	return role, nil
}

func ListAllRoles() ([]map[string]interface{}, error) {
	db := config.DB
	var roles []models.Role
	if err := db.Find(&roles).Error; err != nil {
		return nil, err
	}
	// Convert to a slice of maps
	var result []map[string]interface{}
	for _, roles := range roles {
		result = append(result, gin.H{
			"id":   roles.ID,
			"name": roles.Name,
		})
	}
	return result, nil
}
