package controllers

import (
	"gone-be/config"
	model "gone-be/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(c *gin.Context) {
	// Connect to the database (assuming config.DB is your global DB connection)
	db := config.DB

	// Declare a slice to hold users
	var users []model.User

	// Fetch all users from the database using GORM's Find method
	if err := db.Find(&users).Error; err != nil {
		// If there's an error fetching users, respond with an error message
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users from the database"})
		return
	}

	// Respond with the list of users
	c.JSON(http.StatusOK, users)
}

// POST: Register User
func CreateUser(c *gin.Context) {
	var user model.User

	// Bind the JSON payload to the User struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Basic validation (you can expand this further)
	if user.Username == "" || user.Email == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	// Hash the password before saving it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	// Connect to the database (you should have a global DB connection setup in config)
	db := config.DB

	// Check if user already exists
	var existingUser model.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}

	// Create the user in the database
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Respond with the created user and a 201 status
	c.JSON(http.StatusCreated, user)
}
