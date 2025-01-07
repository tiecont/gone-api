package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"gone-be/config"
	"gone-be/routes"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (adjust for production)
		return true
	},
}

func websocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close WebSocket connection: %v", err)
		}
	}()

	log.Println("WebSocket client connected:", conn.RemoteAddr())

	for {
		// Read message from client
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		log.Printf("Received: %s", string(message))

		// Handle different events
		switch string(message) {
		case "notice":
			response := "Received: " + string(message)
			if err := conn.WriteMessage(messageType, []byte(response)); err != nil {
				log.Printf("Error writing message: %v", err)
				break
			}
		case "bye":
			response := "Disconnected: " + string(message)
			if err := conn.WriteMessage(messageType, []byte(response)); err != nil {
				log.Printf("Error writing message: %v", err)
				break
			}
			break
		default:
			response := "Echo: " + string(message)
			if err := conn.WriteMessage(messageType, []byte(response)); err != nil {
				log.Printf("Error writing message: %v", err)
				break
			}
		}
	}

	log.Println("WebSocket client disconnected:", conn.RemoteAddr())
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
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

	// WebSocket route
	router.GET("/ws", websocketHandler)

	// Register other routes
	routes.RegisterRoutes(router)

	// Connect to the database
	config.ConnectDatabase()

	// Start API and WebSocket server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080" // Default port
	}
	log.Printf("Server running on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
