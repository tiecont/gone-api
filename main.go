package main

import (
	"gone-be/config"
	model "gone-be/models"
	"gone-be/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Tải biến môi trường từ .env
	_ = godotenv.Load()

	// Kết nối cơ sở dữ liệu
	db := config.ConnectDatabase()
	model.MigrateUser(db)
	// Khởi tạo router
	r := gin.Default()

	// Định tuyến (ví dụ)
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Welcome to Gin-Gonic API!"})
	})
	routes.RegisterRoutes(r)
	// Khởi chạy server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server is running on port %s", port)
	r.Run(":" + port)
}
