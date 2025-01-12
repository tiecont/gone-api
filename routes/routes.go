package routes

import (
	"gone-be/modules/user/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.GET("/users", controllers.GetUsers)
		api.POST("/users/register", controllers.RegisterUser)
		api.POST("/users/login", controllers.LoginUser)
	}
}
