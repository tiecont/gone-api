package services

import (
	"errors"
	"gone-be/config"
	model "gone-be/modules/user/models"

	"golang.org/x/crypto/bcrypt"
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

func HandleCreateUser(user model.User) (model.User, error) {
	db := config.DB

	// Kiểm tra dữ liệu đầu vào
	if user.Username == "" || user.Email == "" || user.Password == "" {
		return model.User{}, errors.New("all fields are required")
	}

	// Kiểm tra email đã tồn tại
	var existingUser model.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return model.User{}, errors.New("email already in use")
	}

	// Hash mật khẩu
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, errors.New("failed to hash password")
	}
	user.Password = string(hashedPassword)

	// Tạo người dùng mới trong cơ sở dữ liệu
	if err := db.Create(&user).Error; err != nil {
		return model.User{}, errors.New("failed to create user")
	}

	return user, nil
}
