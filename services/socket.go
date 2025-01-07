package services

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (adjust for production)
		return true
	},
}

// WebSocketHandler handles WebSocket connections and messages
func WebSocketHandler(c *gin.Context) {
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
