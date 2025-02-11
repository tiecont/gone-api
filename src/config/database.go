package config

import (
	"fmt"
	AccessTokens "gone-be/src/modules/key_token/access_token/models"
	RefreshTokens "gone-be/src/modules/key_token/refresh_token/models"
	Permissions "gone-be/src/modules/permission/models"
	"gone-be/src/modules/post/models"
	Profiles "gone-be/src/modules/profile/models"
	Roles "gone-be/src/modules/role/models"
	models2 "gone-be/src/modules/user/models"
	"log"
	"os"

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
		models2.MigrateUser,
		models2.MigrateFriends,
		models2.MigrateMessage,

		// Profile model
		Profiles.MigrateProfile,

		// Permission model
		Permissions.MigratePermissions,
		Roles.MigrateRolePermissions,

		// Access and Refresh Tokens
		AccessTokens.MigrateAccessTokens,
		RefreshTokens.MigrateRefreshTokens,

		// Post-related models
		models.MigratePost,
		models.MigrateComment,
		models.MigrateCommentLike,
		models.MigrateCategory,
		models.MigratePostCategory,
		models.MigratePostLike,
		models.MigrateMedia,
		models.MigratePostShare,

		// Add Tables for future
		models2.MigrateVerificationCode,
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
