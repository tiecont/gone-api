package services

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gone-be/config"
	AccessTokens "gone-be/modules/key_token/access_token/models"
	RefreshTokens "gone-be/modules/key_token/refresh_token/models"
	Profiles "gone-be/modules/profile/models"
	Users "gone-be/modules/user/models"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

func GetUserInfo(accessToken string) (map[string]interface{}, error) {
	db := config.DB

	// Step 1: Validate the access token
	var tokenRecord AccessTokens.AccessToken
	if err := db.Where("token = ?", accessToken).First(&tokenRecord).Error; err != nil {
		return nil, fmt.Errorf("invalid or expired access token")
	}

	// Step 2: Get the user and their associated profile using a raw query
	var results []map[string]interface{} // This will hold the raw query results
	if err := db.Table("users").
		Joins("INNER JOIN profiles ON profiles.user_id = users.id").
		Select("users.id, users.username, users.email, users.phone_number, users.google_id, users.twitter_id, users.password, users.status, users.role_id, users.created_at, users.updated_at, profiles.*").
		Where("users.id = ?", tokenRecord.UserID).
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// The results will now contain the raw data in a map format
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found")
	}

	return results[0], nil
}

// RegisterUser HandleCreateUser handles the logic for creating a new user.
func RegisterUser(user Users.User) (map[string]interface{}, error) {
	db := config.DB

	// Step 1: Validate input data
	if user.Username == "" || user.Email == "" || user.Password == "" {
		return nil, errors.New("all fields (username, email, password) are required")
	}

	// Step 2: Check if the email already exists in the database
	var existingUser Users.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		// Email already exists
		return nil, errors.New("email is already in use")
	}

	// Step 3: Hash the password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		// Error occurred during password hashing
		return nil, errors.New("failed to hash password")
	}
	user.Password = string(hashedPassword)

	// Step 4: Create the new user in the database
	if err := db.Create(&user).Error; err != nil {
		// Error occurred while creating user in the database
		return nil, fmt.Errorf("failed to create user: %v", err.Error())
	}

	// Step 5: Format the response to exclude sensitive data
	response := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"status":     user.Status,
		"role_id":    user.RoleID,
		"twitter_id": user.TwitterID,
		"google_id":  user.GoogleID,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	// Step 6: Return the formatted response
	return response, nil
}

func LoginUser(email, phoneNumber, password, googleToken, twitterID string) (map[string]interface{}, int, error) {
	db := config.DB

	// Step 1: Check for Google Login
	if googleToken != "" {
		result, err := LoginGoogle(googleToken)
		if err != nil {
			return nil, http.StatusUnauthorized, errors.New("invalid Google token")
		}
		return result, http.StatusOK, nil
	}

	// Step 2: Check for Facebook Login
	if twitterID != "" {
		log.Println(twitterID)
	}

	// Step 3: Regular Login (if GoogleID and TwitterID are not provided)
	var existingUser Users.User
	if err := db.Where("email = ? OR phone_number = ?", email, phoneNumber).First(&existingUser).Error; err != nil {
		return nil, http.StatusNotFound, errors.New("invalid email, phone number, or user not found")
	}

	// Step 4: Validate Password
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password)); err != nil {
		return nil, http.StatusUnauthorized, errors.New("invalid password")
	}
	err := DeleteAllTokensByUserID(existingUser.ID)
	if err != nil {
		return nil, 0, err
	}
	// Step 5: Generate Hex Tokens
	accessToken, refreshToken, err := generateHexTokens(existingUser.ID)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("failed to generate tokens")
	}

	// Return the tokens with success status
	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, http.StatusOK, nil
}

// LoginGoogle Google
func LoginGoogle(googleToken string) (map[string]interface{}, error) {
	// Step 1: Send the access token to Google's userinfo endpoint
	userInfoURL := fmt.Sprintf("https://www.googleapis.com/oauth2/v3/userinfo?access_token=%s", googleToken)
	response, err := http.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Google userinfo API: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("failed to close response body")
		}
	}(response.Body)

	// Step 2: Check the response status
	if response.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(response.Body)
		return nil, fmt.Errorf("failed to fetch user info from Google API: %s", string(bodyBytes))
	}

	// Step 3: Parse the JSON response
	var googleUserInfo struct {
		Sub           string `json:"sub"`            // Google user ID
		Email         string `json:"email"`          // User's email
		EmailVerified bool   `json:"email_verified"` // Whether email is verified
		Name          string `json:"name"`           // Full name
		Picture       string `json:"picture"`        // Profile picture URL
	}
	if err := json.NewDecoder(response.Body).Decode(&googleUserInfo); err != nil {
		return nil, fmt.Errorf("failed to decode Google user info: %w", err)
	}

	// Step 4: Check if the user exists in the database
	db := config.DB
	var existingUser Users.User
	if err := db.Where("email = ?", googleUserInfo.Email).First(&existingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// If user doesn't exist, create a new user
			newUser := Users.User{
				GoogleID: googleUserInfo.Sub,
				Username: googleUserInfo.Name,
				Email:    googleUserInfo.Email,
				RoleID:   2, // Default for new user
			}
			if err := db.Create(&newUser).Error; err != nil {
				return nil, fmt.Errorf("failed to create user: %v", err)
			}

			// Create a profile for the new user
			newProfile := Profiles.Profile{
				UserID:     newUser.ID,
				ProfilePic: googleUserInfo.Picture,
				Birthday:   nil,
			}
			if err := db.Create(&newProfile).Error; err != nil {
				return nil, fmt.Errorf("failed to create profile: %v", err)
			}
			existingUser = newUser // Assign the newly created user to `existingUser`
		} else {
			return nil, fmt.Errorf("failed to query user: %v", err)
		}
	} else {
		// User exists, check if Google ID is missing and update it
		if existingUser.GoogleID == "" {
			existingUser.GoogleID = googleUserInfo.Sub
			if err := db.Save(&existingUser).Error; err != nil {
				return nil, fmt.Errorf("failed to update Google ID: %v", err)
			}
		}
	}

	// Step 5: Delete all tokens for the user (cleanup)
	err = DeleteAllTokensByUserID(existingUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete tokens: %w", err)
	}

	// Step 6: Generate hex tokens for the user
	accessToken, refreshToken, err := generateHexTokens(existingUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Step 7: Return the tokens and user info
	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil
}

func LoginTwitter(username, email, image, profileBackgroundImage, twitterId string) (map[string]interface{}, error) {
	db := config.DB

	// Check if the user exists by email
	var existingUser Users.User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// If user doesn't exist, create a new user
			newUser := Users.User{
				TwitterID: twitterId,
				Username:  username,
				Email:     email,
				RoleID:    2, // Default role for new users
			}
			if err := db.Create(&newUser).Error; err != nil {
				return nil, fmt.Errorf("failed to create user: %v", err)
			}

			// Create a profile for the new user
			newProfile := Profiles.Profile{
				UserID:     newUser.ID,
				ProfilePic: image,
				CoverPhoto: profileBackgroundImage,
				Birthday:   nil,
			}
			if err := db.Create(&newProfile).Error; err != nil {
				return nil, fmt.Errorf("failed to create profile: %v", err)
			}

			existingUser = newUser // Assign the newly created user to `existingUser`
		} else {
			// Other errors while querying the database
			return nil, fmt.Errorf("failed to query user: %v", err)
		}
	} else {
		// User exists, update Twitter ID if missing
		if existingUser.TwitterID == "" {
			existingUser.TwitterID = twitterId
		}

		// Update the user's profile fields if they are missing or need updating
		var existingProfile Profiles.Profile
		if err := db.Where("user_id = ?", existingUser.ID).First(&existingProfile).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create a new profile if none exists
				newProfile := Profiles.Profile{
					UserID:     existingUser.ID,
					ProfilePic: image,
					CoverPhoto: profileBackgroundImage,
					Birthday:   nil,
				}
				if err := db.Create(&newProfile).Error; err != nil {
					return nil, fmt.Errorf("failed to create profile: %v", err)
				}
			} else {
				// Other errors while querying the profile
				return nil, fmt.Errorf("failed to query profile: %v", err)
			}
		} else {
			// Update existing profile fields if they are missing or outdated
			if existingProfile.ProfilePic == "" && image != "" {
				existingProfile.ProfilePic = image
			}
			if existingProfile.CoverPhoto == "" && profileBackgroundImage != "" {
				existingProfile.CoverPhoto = profileBackgroundImage
			}
			if err := db.Save(&existingProfile).Error; err != nil {
				return nil, fmt.Errorf("failed to update profile: %v", err)
			}
		}

		// Save the updated user if changes were made
		if err := db.Save(&existingUser).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %v", err)
		}
	}

	// Clean up old tokens
	err := DeleteAllTokensByUserID(existingUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete tokens: %w", err)
	}

	// Generate new tokens
	accessToken, refreshToken, err := generateHexTokens(existingUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Return the tokens and user info
	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil
}

func generateHexTokens(userID uint) (string, string, error) {
	db := config.DB

	// Generate random hex strings for tokens
	accessTokenBytes := make([]byte, 32) // 256-bit
	refreshTokenBytes := make([]byte, 32)

	_, err := rand.Read(accessTokenBytes)
	if err != nil {
		return "", "", errors.New("failed to generate access token")
	}
	_, err = rand.Read(refreshTokenBytes)
	if err != nil {
		return "", "", errors.New("failed to generate refresh token")
	}

	accessToken := hex.EncodeToString(accessTokenBytes)
	refreshToken := hex.EncodeToString(refreshTokenBytes)

	// Save access token to database
	accessTokenEntry := AccessTokens.AccessToken{
		UserID: userID,
		Token:  accessToken,
	}
	if err := db.Create(&accessTokenEntry).Error; err != nil {
		return "", "", err
	}

	// Save refresh token to database
	refreshTokenEntry := RefreshTokens.RefreshToken{
		UserID: userID,
		Token:  refreshToken,
	}
	if err := db.Create(&refreshTokenEntry).Error; err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func DeleteAllTokensByUserID(userID uint) error {
	// Start a transaction to ensure both deletions are atomic
	tx := config.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}

	// Delete all AccessTokens for the given UserID
	if err := tx.Where("user_id = ?", userID).Delete(&AccessTokens.AccessToken{}).Error; err != nil {
		tx.Rollback() // Rollback the transaction if deletion fails
		return fmt.Errorf("failed to delete access tokens: %v", err)
	}

	// Delete all RefreshTokens for the given UserID
	if err := tx.Where("user_id = ?", userID).Delete(&RefreshTokens.RefreshToken{}).Error; err != nil {
		tx.Rollback() // Rollback the transaction if deletion fails
		return fmt.Errorf("failed to delete refresh tokens: %v", err)
	}

	// Commit the transaction if both deletions succeed
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
