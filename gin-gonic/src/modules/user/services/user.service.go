package services

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gone-be/src/config"
	AccessTokens "gone-be/src/modules/key_token/access_token/models"
	RefreshTokens "gone-be/src/modules/key_token/refresh_token/models"
	Profiles "gone-be/src/modules/profile/models"
	Users "gone-be/src/modules/user/models"
	"gone-be/src/utils"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

// GetUserInfo is the function to get user information with the given userID
func GetUserInfo(userID uint) (map[string]interface{}, error) {
	db := config.DB

	// Get the user and their associated profile using a raw query
	var results []map[string]interface{} // This will hold the raw query results
	if err := db.Table("users").
		Joins("INNER JOIN profiles ON profiles.user_id = users.id").
		Select("users.*, profiles.*").
		Where("users.id = ?", userID).Scan(&results).Error; err != nil {
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
	if user.Email == "" || user.Password == "" {
		return nil, errors.New("all fields (email, password) are required")
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

	newUser := Users.User{
		Email:    user.Email,
		Password: user.Password,
		RoleID:   2, // Default for new user
	}
	if err := db.Create(&newUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Step 5: Create a default profile for the new user
	defaultProfile := Profiles.Profile{
		UserID:     newUser.ID,
		ProfilePic: "/default-avatar.png",
		Birthday:   nil, // Default birthday (not set)
	}

	// Step 6: Save the profile to the database
	if err := db.Create(&defaultProfile).Error; err != nil {
		return nil, fmt.Errorf("failed to create profile: %v", err)
	}

	// Step 7: Format the response to exclude sensitive data
	response := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"status":     user.Status,
		"role_id":    user.RoleID,
		"twitter_id": user.TwitterID,
		"google_id":  user.GoogleID,
		"profile": map[string]interface{}{
			"profile_pic":      defaultProfile.ProfilePic,
			"cover_photo":      defaultProfile.CoverPhoto,
			"background_color": defaultProfile.BackgroundColor,
			"sex":              defaultProfile.Gender,
			"birthday":         defaultProfile.Birthday,
			"location":         defaultProfile.Location,
			"bio":              defaultProfile.Bio,
			"created_at":       defaultProfile.CreatedAt,
			"updated_at":       defaultProfile.UpdatedAt,
		},
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	// Step 8: Return the formatted response
	return response, nil
}

// LoginUser is the function to help user login and return the tokens
func LoginUser(email, phoneNumber, password, username string) (string, int, error) {
	db := config.DB

	// Step 1: Xác định thông tin đầu vào để truy vấn
	var existingUser Users.User
	var err error

	switch {
	case email != "":
		err = db.Where("status = true AND email = ?", email).First(&existingUser).Error
	case phoneNumber != "":
		err = db.Where("status = true AND phone_number = ?", phoneNumber).First(&existingUser).Error
	case username != "":
		err = db.Where("status = true AND username = ?", username).First(&existingUser).Error
	default:
		return "", http.StatusBadRequest, errors.New("email, phone number, or username is required")
	}

	if err != nil {
		return "", http.StatusUnauthorized, errors.New("invalid user")
	}

	// Step 2: Kiểm tra mật khẩu
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password)); err != nil {
		return "", http.StatusUnauthorized, errors.New("invalid credentials")
	}

	// Step 3: Tạo mã xác minh 6 chữ số
	verificationCode := generateVerificationCode()

	// Step 4: Lưu mã xác minh
	if err := saveVerificationCode(existingUser.Email, verificationCode); err != nil {
		return "", http.StatusInternalServerError, errors.New("failed to save verification code")
	}

	// Step 5: Gửi mã xác minh qua email
	if err := sendVerificationEmail(existingUser.Email, verificationCode); err != nil {
		return "", http.StatusInternalServerError, errors.New(err.Error())
	}

	// Step 6: Trả về thành công
	return existingUser.Email, http.StatusOK, nil
}

// LoginGoogle is the function to login by google service and return the tokens
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

// LoginTwitter is the function to help user login via twitter service and return the tokens
func LoginTwitter(username, email, image, profileBackgroundImage, profileBackgroundColor, twitterId string) (map[string]interface{}, error) {
	db := config.DB

	// Check if the user exists by email
	var existingUser Users.User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
				UserID:          newUser.ID,
				ProfilePic:      image,
				CoverPhoto:      profileBackgroundImage,
				BackgroundColor: profileBackgroundColor,
				Birthday:        nil,
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
					UserID:          existingUser.ID,
					ProfilePic:      image,
					CoverPhoto:      profileBackgroundImage,
					BackgroundColor: profileBackgroundColor,
					Birthday:        nil,
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
				existingProfile.BackgroundColor = profileBackgroundColor
			}
			if err := db.Save(&existingProfile).Error; err != nil {
				return nil, fmt.Errorf("failed to update profile: %v", err)
			}
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

// RefreshAccessToken is the function to renew the access token by refresh token
func RefreshAccessToken(token, clientID string) (map[string]interface{}, error) {
	tx := config.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}

	// Check if the refresh token is valid
	var refreshToken RefreshTokens.RefreshToken
	if err := tx.Where("token = ? AND user_id = ?", token, clientID).First(&refreshToken).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("invalid or expired refresh token: %v", err)
	}

	// Delete old access tokens for the user
	if err := tx.Where("user_id = ?", refreshToken.UserID).Delete(&AccessTokens.AccessToken{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete old access tokens: %v", err)
	}

	// Generate a new access token
	token, err := GenerateAccessToken()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to generate new access token: %v", err)
	}

	newAccessToken := AccessTokens.AccessToken{
		Token:  token,
		UserID: refreshToken.UserID,
		Status: true,
	}

	if err := tx.Create(&newAccessToken).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create new access token: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Return the new access token
	response := map[string]interface{}{
		"access_token": newAccessToken.Token,
		"expires_at":   newAccessToken.ExpiresAt,
	}

	return response, nil
}

// GenerateAccessToken is the function to generate an access token
func GenerateAccessToken() (string, error) {
	// Generate random bytes for the token
	tokenBytes := make([]byte, 32) // 256-bit token
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate access token: %v", err)
	}

	// Encode the random bytes to a hexadecimal string
	accessToken := hex.EncodeToString(tokenBytes)

	return accessToken, nil
}

// ForgotPassword is the function will receive an email to process forgot password service
func ForgotPassword(email string) (int, error) {
	db := config.DB
	// Step 1: Check user?
	var user Users.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return http.StatusBadRequest, errors.New("user not found")
	}
	// Step 2: Generate 6 digits code
	verificationCode := generateVerificationCode()

	// Step 3: Save the verification code to the database
	if err := saveVerificationCode(email, verificationCode); err != nil {
		return http.StatusInternalServerError, errors.New("failed to save verification code")
	}

	// Step 4: Send the verification code via email
	if err := sendVerificationEmail(email, verificationCode); err != nil {
		return http.StatusInternalServerError, errors.New("failed to send verification email")
	}

	return http.StatusOK, nil
}

// RenewPassword is the function to change the password follow by user
func RenewPassword(newPassword, userID string) (int, error) {
	// Start a database transaction
	tx := config.DB.Begin()
	var user Users.User

	// Step 1: Check if the user exists
	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		tx.Rollback()
		return http.StatusNotFound, errors.New("user not found")
	}

	// Step 2: Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, errors.New("failed to hash new password")
	}

	// Step 4: Update the password in the database
	user.Password = string(hashedPassword)
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, errors.New("failed to save new password")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	// Step 5: Return success response
	return http.StatusOK, nil
}

// ChangePassword is the function to change the password follow by user
func ChangePassword(oldPassword, newPassword, userID string) (int, error) {
	// Start a database transaction
	tx := config.DB.Begin()
	var user Users.User

	// Step 1: Check if the user exists
	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		tx.Rollback()
		return http.StatusNotFound, errors.New("user not found")
	}

	// Step 2: Validate the old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		tx.Rollback()
		return http.StatusUnauthorized, errors.New("old password is incorrect")
	}

	// Step 3: Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, errors.New("failed to hash new password")
	}

	// Step 4: Update the password in the database
	user.Password = string(hashedPassword)
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, errors.New("failed to save new password")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return http.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	// Step 5: Return success response
	return http.StatusOK, nil
}

// GetUserImageByID is a function to get user avatar by ID, this function will open for all clients
// GetUserImageByID retrieves a user's profile picture URL by their user ID
func GetUserImageByID(userID uint) (string, *utils.ServiceError) {
	if userID == 0 {
		return "", &utils.ServiceError{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid user ID",
		}
	}

	var userAvatar string
	err := config.DB.
		Model(&Profiles.Profile{}).
		Select("profile_pic").
		Where("user_id = ?", userID).
		Scan(&userAvatar).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", &utils.ServiceError{
				StatusCode: http.StatusNotFound,
				Message:    "user profile not found",
			}
		}

		return "", &utils.ServiceError{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("failed to get user image: %v", err),
		}
	}

	// Return empty string if no profile picture is set
	return userAvatar, nil
}
