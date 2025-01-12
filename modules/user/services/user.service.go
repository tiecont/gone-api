package services

import (
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gone-be/config"
	AccessTokens "gone-be/modules/key_token/access_token/models"
	RefreshTokens "gone-be/modules/key_token/refresh_token/models"
	Users "gone-be/modules/user/models"
	"math/rand"
	"net/http"
)

func GetAllUsers() ([]Users.User, error) {
	db := config.DB
	var users []Users.User

	// Find users
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
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
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"status":      user.Status,
		"role_id":     user.RoleID,
		"facebook_id": user.FacebookID,
		"google_id":   user.GoogleID,
		"created_at":  user.CreatedAt,
		"updated_at":  user.UpdatedAt,
	}

	// Step 6: Return the formatted response
	return response, nil
}

func LoginUser(email, phoneNumber, password, googleToken, facebookToken string) (map[string]interface{}, int, error) {
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
	if facebookToken != "" {
		result, err := LoginFacebook(facebookToken)
		if err != nil {
			return nil, http.StatusUnauthorized, errors.New("invalid Facebook token")
		}
		return result, http.StatusOK, nil
	}

	// Step 3: Regular Login (if GoogleID and FacebookID are not provided)
	var existingUser Users.User
	if err := db.Where("email = ? OR phone_number = ?", email, phoneNumber).First(&existingUser).Error; err != nil {
		return nil, http.StatusNotFound, errors.New("invalid email, phone number, or user not found")
	}

	// Step 4: Validate Password
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password)); err != nil {
		return nil, http.StatusUnauthorized, errors.New("invalid password")
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
	// Simulate Google token validation (you should replace this with an actual Google API validation)
	if googleToken == "" {
		return nil, errors.New("invalid Google token")
	}

	// Assuming a valid user ID is retrieved from the Google token
	userID := uint(1) // Example user ID

	// Generate JWT tokens
	accessToken, refreshToken, err := generateHexTokens(userID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, nil
}

// LoginFacebook login by facebook
func LoginFacebook(facebookToken string) (map[string]interface{}, error) {
	// Simulate Facebook token validation (you should replace this with an actual Facebook API validation)
	if facebookToken == "" {
		return nil, errors.New("invalid Facebook token")
	}

	// Assuming a valid user ID is retrieved from the Facebook token
	userID := uint(2) // Example user ID

	// Generate JWT tokens
	accessToken, refreshToken, err := generateHexTokens(userID)
	if err != nil {
		return nil, err
	}

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
