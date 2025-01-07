package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/joho/godotenv"
	"gone-be/config"
	"gone-be/routes"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Setup Gin router
	router := gin.Default()

	// Enable CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Setup Socket.IO server
	server := socketio.NewServer(nil)

	// Socket.IO events
	server.OnConnect("/", func(s socketio.Conn) error {
		log.Println("Client connected, ID:", s.ID())
		s.Emit("heartbeat", "Connection established") // Emit initial connection message
		return nil
	})

	server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
		log.Println("Notice event received:", msg)
		s.Emit("reply", "Received: "+msg)
	})

	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		log.Println("Chat message received:", msg)
		s.SetContext(msg) // Store context if needed
		return "Message received: " + msg
	})

	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		log.Println("Bye event received")
		var lastMsg string
		if s.Context() != nil {
			lastMsg = s.Context().(string)
		} else {
			lastMsg = "No previous message"
		}
		s.Emit("bye", lastMsg)
		s.Close()
		return "Disconnected: " + lastMsg
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Printf("Client disconnected (ID: %s), Reason: %s\n", s.ID(), reason)
	})

	server.OnError("/", func(s socketio.Conn, err error) {
		log.Printf("Socket error (ID: %s): %v\n", s.ID(), err.Error())
		s.Emit("error", err.Error())
	})

	// Start the Socket.IO server in a goroutine
	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("Socket.IO server failed: %v\n", err)
		}
	}()
	defer server.Close()

	// Integrate Socket.IO with Gin
	router.GET("/socket.io/*any", gin.WrapH(server))
	router.POST("/socket.io/*any", gin.WrapH(server))

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
