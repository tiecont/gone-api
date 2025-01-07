package services

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gone-be/config"
	model "gone-be/modules/user/models"
)

func GetAllUsers() ([]model.User, error) {
	db := config.DB
	var users []model.User

	// Lấy danh sách người dùng từ cơ sở dữ liệu
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// HandleCreateUser handles the logic for creating a new user.
func HandleCreateUser(user model.User) (model.User, error) {
	db := config.DB

	// Step 1: Validate input data
	if user.Username == "" || user.Email == "" || user.Password == "" {
		return model.User{}, errors.New("all fields (Username, Email, Password) are required")
	}

	// Step 2: Check if the email already exists in the database
	var existingUser model.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		// Email already exists
		return model.User{}, errors.New("email is already in use")
	}

	// Step 3: Hash the password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		// Error occurred during password hashing
		return model.User{}, errors.New("failed to hash password")
	}
	user.Password = string(hashedPassword)

	// Step 4: Create the new user in the database
	if err := db.Create(&user).Error; err != nil {
		// Error occurred while creating user in the database, formatted with err.Error()
		return model.User{}, fmt.Errorf("failed to create user: %v", err.Error())
	}

	// Step 5: Return the created user
	return user, nil
}
