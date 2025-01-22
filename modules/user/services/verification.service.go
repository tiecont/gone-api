package services

import (
	"fmt"
	"gone-be/config"
	AccessTokens "gone-be/modules/key_token/access_token/models"
	Users "gone-be/modules/user/models"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"time"
)

// Helper function to generate a 6-digit verification code
func generateVerificationCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// Helper function to save the verification code
func saveVerificationCode(email, code string) error {
	db := config.DB
	var user Users.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return err
	}
	// Delete all verification code before sent a new one
	if err := db.Where("email = ?", email).Delete(&Users.VerificationCode{}).Error; err != nil {
		return fmt.Errorf("failed to delete verification code: %v", err)
	}

	verification := Users.VerificationCode{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(3 * time.Minute),
	}
	return db.Create(&verification).Error
}

// sendVerificationEmail sends a verification email with the given code
func sendVerificationEmail(email, code string) error {
	// Email server configuration
	smtpHost := os.Getenv("SMTP_HOST")     // E.g., "smtp.gmail.com"
	smtpPort := os.Getenv("SMTP_PORT")     // E.g., "587"
	smtpUsername := os.Getenv("SMTP_USER") // Your email address
	smtpPassword := os.Getenv("SMTP_PASS") // Your email password or app-specific password

	// Email content
	subject := "Your Verification Code"
	body := fmt.Sprintf("Your verification code is: %s", code)
	message := fmt.Sprintf("Subject: %s\n\n%s", subject, body)

	// Set up authentication information
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	// Send the email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", smtpHost, smtpPort),
		auth,
		smtpUsername,
		[]string{email},
		[]byte(message),
	)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// VerifyCodeAndGenerateTokens Function to verify the code and generate tokens
func VerifyCodeAndGenerateTokens(code Users.VerificationCode) (map[string]interface{}, int, error) {
	db := config.DB
	// Step 1: Retrieve the user associated with the email
	var user Users.User
	if err := db.Where("email = ?", code.Email).First(&user).Error; err != nil {
		return nil, http.StatusNotFound, fmt.Errorf("invalid user")
	}

	// Step 2: Retrieve the verification record
	var verification Users.VerificationCode
	if err := db.Where("email = ?", code.Email).First(&verification).Error; err != nil {
		return nil, http.StatusNotFound, fmt.Errorf("verification record not found")
	}

	// Step 3: Check if the code is expired
	if verification.ExpiresAt.Before(time.Now()) {
		db.Delete(&verification) // Ensure to delete expired verification code
		return nil, http.StatusUnauthorized, fmt.Errorf("verification code has expired")
	}

	// Step 4: Check valid code
	if verification.Code != code.Code {
		verification.InputCount += 1
		db.Save(&verification)

		// Step 5: Lock user after 5 failed attempts
		if verification.InputCount >= 5 {
			db.Delete(&verification)
			user.Status = false
			db.Save(&user)
			return nil, http.StatusUnauthorized, fmt.Errorf("too many time argument, your account had been suspended, please get contact to admin")
		}
		return nil, http.StatusUnauthorized, fmt.Errorf("invalid verification code")
	}

	// Step 6: Delete all existing tokens for the user
	if err := DeleteAllTokensByUserID(user.ID); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to delete existing tokens")
	}

	// Step 7: Generate new tokens
	accessToken, refreshToken, err := generateHexTokens(user.ID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to generate tokens")
	}

	// Step 8: Return the tokens with success status
	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, http.StatusOK, nil
}

func VerifyCode(code, email string) (map[string]interface{}, error) {
	// Import db queries
	db := config.DB
	var user Users.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	var verification Users.VerificationCode
	if err := db.Where("email = ? AND code = ?", email, code).First(&verification).Error; err != nil {
		return nil, err
	}
	// Step 3: Check if the code is expired
	if verification.ExpiresAt.Before(time.Now()) {
		db.Delete(&verification) // Ensure to delete expired verification code
		return nil, fmt.Errorf("verification code has expired")
	}
	// Step 4: Check valid code
	if verification.Code != code {
		verification.InputCount += 1
		db.Save(&verification)

		// Step 5: Lock user after 5 failed attempts
		if verification.InputCount >= 5 {
			db.Delete(&verification)
			user.Status = false
			db.Save(&user)
			return nil, fmt.Errorf("too many time argument, your account had been suspended, please get contact to admin")
		}
		return nil, fmt.Errorf("invalid verification code")
	}
	// Step 6: Delete all existing tokens for the user
	if err := DeleteAllTokensByUserID(user.ID); err != nil {
		return nil, fmt.Errorf("failed to delete existing tokens")
	}

	// Step 7: Get New Token
	token, err := GenerateAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token")
	}
	// Step 8: Save token to user
	accessTokenEntry := AccessTokens.AccessToken{
		UserID: user.ID,
		Token:  token,
	}
	if err := db.Create(&accessTokenEntry).Error; err != nil {
		return nil, fmt.Errorf("failed to save access token")
	}

	// Step 9: Return the token and user ID
	response := map[string]interface{}{
		"token":   accessTokenEntry.Token,
		"user_id": accessTokenEntry.UserID,
	}
	return response, nil
}
