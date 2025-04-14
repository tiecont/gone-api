package main

import (
	"fmt"
	"gone-be/src/config"
	"gone-be/src/routes"
	"gone-be/src/services"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("⚠️ Could not load .env file: %v", err)
	} else {
		log.Println("✅ .env file loaded successfully")
	}

	env := os.Getenv("NODE_ENV")
	log.Printf("NODE_ENV = %s", env)

	host := os.Getenv("HOST")
	port := os.Getenv("APP_PORT")
	if host == "" {
		host = "0.0.0.0"
	}
	if port == "" {
		port = "2000"
	}
	// Setup Gin router
	router := gin.Default()
	// Enable CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Set static files
	router.Static("/uploads", "./uploads")

	// WebSocket route
	router.GET("/ws", services.WebSocketHandler)

	// Connect to the database
	config.ConnectDatabase()

	// Register other routes
	routes.RegisterRoutes(router)

	// Start API and WebSocket server
	addr := fmt.Sprintf("%s:%s", host, port)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
