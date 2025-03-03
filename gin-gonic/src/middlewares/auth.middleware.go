package middlewares

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gone-be/src/modules/key_token/access_token/services"
	PermissionServices "gone-be/src/modules/permission/services"
	"gone-be/src/modules/user/models"
	"net/http"
	"strings"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization token missing"})
			c.Abort()
			return
		}

		token, err := SplitToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": err.Error(),
			})
			return
		}

		userInfo, err := services.VerifyToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			c.Abort()
			return
		}

		// Store user in context for later use
		c.Set("user", userInfo)
		c.Next()
	}
}

// Authorization ensures the user has the required permission
func Authorization(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve user from context
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Type assertion
		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			c.Abort()
			return
		}

		// Check role permissions
		if !PermissionServices.CheckPermission(user.RoleID, requiredPermission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func SplitToken(token string) (string, error) {
	splitToken := strings.Split(token, " ")
	// Check if the token is in the correct "Bearer <token>" format
	if len(splitToken) != 2 || strings.ToLower(splitToken[0]) != "bearer" {
		return "", errors.New("token format error")
	}
	return splitToken[1], nil
}
