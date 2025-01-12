package config

import (
	"fmt"
	"log"
	"os"

	AccessTokens "gone-be/modules/key_token/access_token/models"
	RefreshTokens "gone-be/modules/key_token/refresh_token/models"
	Permissions "gone-be/modules/permission/models"
	Posts "gone-be/modules/post/models"
	Profiles "gone-be/modules/profile/models"
	Roles "gone-be/modules/role/models"
	Users "gone-be/modules/user/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDatabase initializes and migrates the database.
func ConnectDatabase() *gorm.DB {
	// Retrieve connection info from environment variables
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Format MySQL DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// Connect to MySQL
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Enable SQL query logs
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = database
	fmt.Println("Connected to database!")

	// Perform database migrations
	if err := runMigrations(DB); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	return DB
}

// runMigrations runs all model migrations.
func runMigrations(db *gorm.DB) error {
	migrations := []func(*gorm.DB) error{
		// Role and User models
		Roles.MigrateRole,
		Users.MigrateUser,
		Users.MigrateFriends,
		Users.MigrateMessage,

		// Profile model
		Profiles.MigrateProfile,

		// Permission model
		Permissions.MigratePermissions,

		// Access and Refresh Tokens
		AccessTokens.MigrateAccessTokens,
		RefreshTokens.MigrateRefreshTokens,

		// Post-related models
		Posts.MigratePost,
		Posts.MigrateComment,
		Posts.MigrateCommentLike,
		Posts.MigrateCategory,
		Posts.MigratePostCategory,
		Posts.MigratePostLike,
		Posts.MigrateMedia,
		Posts.MigratePostShare,
	}

	// Iterate through all migrations
	for _, migrate := range migrations {
		if err := migrate(db); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	
	fmt.Println("All migrations completed successfully!")
	return nil
}
