package routes

import (
	"gone-be/modules/user/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.GET("/auth/user-info", controllers.GetUser)
		api.POST("/users/register", controllers.RegisterUser)
		api.POST("/auth/login", controllers.LoginUser)
		api.POST("/auth/login/google", controllers.LoginGoogle)
		api.POST("/auth/login/twitter", controllers.LoginTwitter)
	}
}
