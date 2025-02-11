package services

import (
	"errors"
	"gone-be/src/config"
	"gone-be/src/modules/permission/models"
	Role "gone-be/src/modules/role/models"
	RoleServices "gone-be/src/modules/role/services"
	"gorm.io/gorm"
)

// CheckPermission verifies if a role has a specific permission
func CheckPermission(roleID uint, permissionName string) bool {
	db := config.DB

	var permission models.Permission
	if err := db.Where("name = ?", permissionName).First(&permission).Error; err != nil {
		return false
	}
	var exists bool
	err := db.Model(&Role.RolePermission{}).
		Select("1").
		Where("role_id = ? AND permission_id = ?", roleID, permission.ID).
		Limit(1).
		Scan(&exists).Error

	if err != nil || !exists {
		return false
	}

	return exists
}

func CreatePermission(permissionName string) (bool, error) {
	db := config.DB

	var exists bool
	err := db.Model(models.Permission{}).Select("name").Where("name = ?", permissionName).First(&models.Permission{}).Scan(&exists).Error
	if err == nil || exists {
		return false, errors.New("permission already exists")
	}

	var permission models.Permission
	permission.Name = permissionName

	if err := db.Create(&permission).Error; err != nil {
		return false, errors.New("failed to create permission")
	}
	return true, nil
}

func ListAllPermissions() ([]map[string]interface{}, error) {
	db := config.DB

	// Fetch all roles
	roles, err := RoleServices.ListAllRoles()
	if err != nil {
		return nil, err
	}

	var roleIDs []uint
	for _, role := range roles {
		roleIDs = append(roleIDs, role["id"].(uint)) // Extracting role IDs
	}

	// Fetch permissions grouped by role_id
	var rolePermissions []struct {
		RoleID         uint
		PermissionID   uint
		PermissionName string
		RoleName       string
	}

	if err := db.Table("role_permissions").
		Select("role_permissions.role_id, permissions.id as permission_id, permissions.name as permission_name, roles.name as role_name").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN roles ON role_permissions.role_id = roles.id").
		Where("role_permissions.role_id IN ?", roleIDs).
		Scan(&rolePermissions).Error; err != nil {
		return nil, err
	}

	// Organizing data into role-wise permission mapping
	rolePermissionsMap := make(map[uint]map[string]interface{})

	for _, rp := range rolePermissions {
		if _, exists := rolePermissionsMap[rp.RoleID]; !exists {
			rolePermissionsMap[rp.RoleID] = map[string]interface{}{
				"role_id":     rp.RoleID,
				"role_name":   rp.RoleName,
				"permissions": []map[string]interface{}{},
			}
		}

		rolePermissionsMap[rp.RoleID]["permissions"] = append(
			rolePermissionsMap[rp.RoleID]["permissions"].([]map[string]interface{}),
			map[string]interface{}{
				"id":   rp.PermissionID,
				"name": rp.PermissionName,
			},
		)
	}

	// Convert map to slice for final output
	var result []map[string]interface{}
	for _, value := range rolePermissionsMap {
		result = append(result, value)
	}

	return result, nil
}

func AddPermissionsToRole(roleID uint, permissionIDs []uint) (map[string]interface{}, error) {
	db := config.DB

	// 1. Check if the Role exists
	var role Role.Role
	if err := db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	// 2. Delete existing permissions for the role in role_permissions
	if err := db.Where("role_id = ?", roleID).Delete(&Role.RolePermission{}).Error; err != nil {
		return nil, errors.New("failed to remove old permissions")
	}

	// 3. Check if all provided Permission IDs exist
	var permissions []models.Permission
	if err := db.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
		return nil, err
	}

	// Ensure all provided permission IDs exist in the system
	if len(permissions) != len(permissionIDs) {
		return nil, errors.New("some permissions not found")
	}

	// 4. Assign new permissions to the role
	var rolePermissions []Role.RolePermission
	for _, permission := range permissions {
		rolePermissions = append(rolePermissions, Role.RolePermission{
			RoleID:       roleID,
			PermissionID: permission.ID,
		})
	}

	// 5. Bulk insert into role_permissions table
	if err := db.Create(&rolePermissions).Error; err != nil {
		return nil, err
	}

	// 6. Prepare response data
	response := map[string]interface{}{
		"role_id":   role.ID,
		"role_name": role.Name,
		"permissions": func() []map[string]interface{} {
			var perms []map[string]interface{}
			for _, perm := range permissions {
				perms = append(perms, map[string]interface{}{
					"id":   perm.ID,
					"name": perm.Name,
				})
			}
			return perms
		}(),
	}

	return response, nil
}
