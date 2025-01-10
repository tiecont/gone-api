package services

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Upgrade configuration for WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins (adjust for production)
		return true
	},
}

// WebSocketMessage represents the JSON message format
type WebSocketMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// WebSocketHandler handles WebSocket connections and events
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

	log.Printf("Client connected: %s", conn.RemoteAddr())

	for {
		// Read message from the client
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Parse the JSON message
		var wsMessage WebSocketMessage
		if err := json.Unmarshal(message, &wsMessage); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		// Process the message
		handleMessage(conn, messageType, wsMessage)
	}

	log.Printf("Client disconnected: %s", conn.RemoteAddr())
}

// handleMessage processes incoming WebSocket messages
func handleMessage(conn *websocket.Conn, messageType int, wsMessage WebSocketMessage) {
	var response WebSocketMessage

	switch wsMessage.Type {
	case "notice":
		response = WebSocketMessage{
			Type:    "notice",
			Message: "Received: " + wsMessage.Message,
		}
	case "bye":
		response = WebSocketMessage{
			Type:    "bye",
			Message: "Goodbye: " + wsMessage.Message,
		}
	case "ping":
		response = WebSocketMessage{
			Type:    "ping",
			Message: "Pong: " + wsMessage.Message,
		}
	default:
		response = WebSocketMessage{
			Type:    "echo",
			Message: "Echo: " + wsMessage.Message,
		}
	}

	// Send the response as JSON
	if err := sendJSONMessage(conn, messageType, response); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// sendJSONMessage sends a JSON-encoded message to the client
func sendJSONMessage(conn *websocket.Conn, messageType int, response WebSocketMessage) error {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return err
	}
	return conn.WriteMessage(messageType, responseJSON)
}
