package services

import (
	"encoding/json"
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

// WebSocketMessage Define a struct to represent the JSON message format
type WebSocketMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
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

	for {
		// Read message from client
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Unmarshal the JSON message into the WebSocketMessage struct
		var wsMessage WebSocketMessage
		err = json.Unmarshal(message, &wsMessage)
		if err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			break
		}

		// Handle events using switch-case
		switch wsMessage.Type {
		case "notice":
			response := WebSocketMessage{
				Type:    "notice",
				Message: "Received: " + wsMessage.Message,
			}
			// Marshal response to JSON
			responseJSON, err := json.Marshal(response)
			if err != nil {
				log.Printf("Error marshalling response: %v", err)
				break
			}
			// Send the JSON response
			if err := conn.WriteMessage(messageType, responseJSON); err != nil {
				log.Printf("Error writing message: %v", err)
				break
			}
		case "bye":
			response := WebSocketMessage{
				Type:    "bye",
				Message: "Disconnected: " + wsMessage.Message,
			}
			// Marshal response to JSON
			responseJSON, err := json.Marshal(response)
			if err != nil {
				log.Printf("Error marshalling response: %v", err)
				break
			}
			// Send the JSON response
			if err := conn.WriteMessage(messageType, responseJSON); err != nil {
				log.Printf("Error writing message: %v", err)
				break
			}
			break
		case "ping":
			// Send a "pong" response as a JSON string
			response := WebSocketMessage{
				Type:    "ping",
				Message: "pong",
			}
			// Marshal response to JSON
			responseJSON, err := json.Marshal(response)
			if err != nil {
				log.Printf("Error marshalling response: %v", err)
				break
			}
			// Send the JSON response
			if err := conn.WriteMessage(websocket.TextMessage, responseJSON); err != nil {
				log.Printf("Error writing message: %v", err)
				break
			}
		default:
			response := WebSocketMessage{
				Type:    "echo",
				Message: wsMessage.Message,
			}
			// Marshal response to JSON
			responseJSON, err := json.Marshal(response)
			if err != nil {
				log.Printf("Error marshalling response: %v", err)
				break
			}
			// Send the JSON response
			if err := conn.WriteMessage(messageType, responseJSON); err != nil {
				log.Printf("Error writing message: %v", err)
				break
			}
		}
	}

	log.Println("WebSocket client disconnected:", conn.RemoteAddr())
}
