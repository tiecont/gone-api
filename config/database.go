package config

import (
	"fmt"
	Roles "gone-be/modules/role/models"
	Users "gone-be/modules/user/models"
	"gorm.io/gorm/logger"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() *gorm.DB {
	// Lấy thông tin kết nối từ biến môi trường
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Chuỗi kết nối MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// Kết nối MySQL
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Enable SQL query logs
	})
	// Gán DB toàn cục
	DB = database
	// Enable query logs for debugging
	DB = DB.Session(&gorm.Session{
		Logger: logger.Default.LogMode(logger.Info),
	})
	err = Roles.MigrateRole(DB) // Migrate Role first
	if err != nil {
		log.Fatalf("Failed to migrate Role: %v", err)
	}

	err = Users.MigrateUser(DB) // Then migrate User
	if err != nil {
		log.Fatalf("Failed to migrate User: %v", err)
	}

	fmt.Println("Connected to database!")
	return DB
}
