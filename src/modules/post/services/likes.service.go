package services

import (
	"fmt"
	"gone-be/src/functions"
	"gone-be/src/utils"
)

// Like is a service function to process the like controller
func Like(userID, postID uint) *utils.ServiceError {

	testResults, err := functions.TestGraphQL()
	if err != nil {
		return &utils.ServiceError{
			StatusCode: err.StatusCode,
			Message:    err.Message,
		}
	}
	fmt.Println(testResults)
	return nil
}
