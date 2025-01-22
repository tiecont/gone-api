package routes

import (
	UserControler "gone-be/modules/user/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// Authentication
		api.GET("/auth/user-info", UserControler.GetUser)                        // 1
		api.POST("/auth/register", UserControler.RegisterUser)                   // 2
		api.POST("/auth/login", UserControler.LoginUser)                         // 3
		api.POST("/auth/login/google", UserControler.LoginGoogle)                // 4
		api.POST("/auth/login/twitter", UserControler.LoginTwitter)              // 5
		api.POST("/auth/verification", UserControler.VerifyCodeAndGenerateToken) // 6
		api.POST("/auth/refresh-access-token", UserControler.RefreshAccessToken) // 7
		api.POST("/auth/forgot-password", UserControler.ForgotPassword)          // 8
		api.POST("/auth/confirm-forgot-password", UserControler.VerifyCode)      // 9
		api.POST("/auth/change-password", UserControler.ChangePassword)          // 10

		// Main application APIs
	}
}
