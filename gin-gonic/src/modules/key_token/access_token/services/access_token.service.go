package services

import (
	"gone-be/src/config"
	AccessTokens "gone-be/src/modules/key_token/access_token/models"
	Users "gone-be/src/modules/user/models"
)

func VerifyToken(token string) (*Users.User, error) {
	db := config.DB
	// Proceed with token validation
	var tokenRecord AccessTokens.AccessToken
	if err := db.Where("token = ?", token).First(&tokenRecord).Error; err != nil {
		return nil, err
	}

	// Find User By ID
	var user Users.User
	if err := db.Where("id = ?", tokenRecord.UserID).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
