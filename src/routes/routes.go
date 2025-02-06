package routes

import (
	"gone-be/src/middlewares"
	PermissionController "gone-be/src/modules/permission/controllers"
	PostControllers "gone-be/src/modules/post/controllers"
	RoleControllers "gone-be/src/modules/role/controllers"
	UserControllers "gone-be/src/modules/user/controllers"
	"gone-be/src/utils"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes initializes all API routes
func RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")

	// Middlewares
	authMiddleware := middlewares.AuthMiddleware
	authzMiddleware := middlewares.Authorization
	needPermission := utils.Permissions
	// Authentication Routes
	authRoutes := api.Group("/auth")
	{
		authRoutes.GET("user-info", authMiddleware(), authzMiddleware(needPermission["ALLOW_GET_USER"]), UserControllers.GetUser) // 1
		authRoutes.POST("register", UserControllers.RegisterUser)                                                                 // 2
		authRoutes.POST("login", UserControllers.LoginUser)                                                                       // 3
		authRoutes.POST("login/google", UserControllers.LoginGoogle)                                                              // 4
		authRoutes.POST("login/twitter", UserControllers.LoginTwitter)                                                            // 5
		authRoutes.POST("verification", UserControllers.VerifyCodeAndGenerateToken)                                               // 6
		authRoutes.POST("refresh-access-token", UserControllers.RefreshAccessToken)                                               // 7
		authRoutes.POST("forgot-password", UserControllers.ForgotPassword)                                                        // 8
		authRoutes.POST("confirm-forgot-password", UserControllers.VerifyCode)                                                    // 9
		authRoutes.POST("change-password", UserControllers.ChangePassword)                                                        // 10
	}

	// Main Application Routes
	// Post Group APIs
	postsRoutes := api.Group("/posts")
	{
		postsRoutes.GET("categories", authMiddleware(), authzMiddleware(needPermission["ALLOW_LIST_CATEGORIES"]), PostControllers.ListCategories) // 11
		postsRoutes.POST("create", authMiddleware(), authzMiddleware(needPermission["ALLOW_CREATE_POST"]), PostControllers.CreatePost)            // 12
	}

	// Role Group APIs
	rolesRoutes := api.Group("/roles")
	{
		rolesRoutes.POST("list", authMiddleware(), authzMiddleware(needPermission["ALLOW_LIST_ROLES"]), RoleControllers.ListRoles)     // 13
		rolesRoutes.POST("create", authMiddleware(), authzMiddleware(needPermission["ALLOW_CREATE_ROLE"]), RoleControllers.CreateRole) // 14
	}
	// Permission Group APIs
	permissionsRoutes := api.Group("/permissions")
	{
		permissionsRoutes.POST("list", authMiddleware(), authzMiddleware(needPermission["ALLOW_LIST_PERMISSIONS"]), PermissionController.ListPermissions)             // 15
		permissionsRoutes.POST("create", authMiddleware(), authzMiddleware(needPermission["ALLOW_CREATE_PERMISSION"]), PermissionController.CreatePermission)         // 16
		permissionsRoutes.POST("assign", authMiddleware(), authzMiddleware(needPermission["ALLOW_ASSIGN_PERMISSIONS"]), PermissionController.AssignPermissionsToRole) // 17
	}
}
