package functions

import (
	"github.com/gin-gonic/gin"
	model "gone-be/src/modules/user/models"
	"gone-be/src/utils"
	"net/http"
)

// GetCurrentUser retrieves the current user from the context
func GetCurrentUser(c *gin.Context) (*model.User, *utils.ServiceError) {
	user, exists := c.Get("user")
	if !exists {
		return nil, &utils.ServiceError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Unauthorized: user not found",
		}
	}

	// Type assertion (ensure user is of type *model.User)
	currentUser, ok := user.(*model.User)
	if !ok {
		return nil, &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Type assertion failed: user is not of type *model.User",
		}
	}

	return currentUser, nil
}
