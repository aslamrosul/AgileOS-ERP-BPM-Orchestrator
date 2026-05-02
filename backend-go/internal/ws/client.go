package ws

import (
	"encoding/json"
	"net/http"
	"time"

	"agileos-backend/auth"
	"agileos-backend/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should validate the origin
		return true
	},
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// The hub
	hub *Hub

	// User ID
	UserID string

	// Client ID for tracking
	ID string

	// User information
	Username string
	Role     string
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.LogError("WebSocket error", err, map[string]interface{}{
					"user_id":   c.UserID,
					"client_id": c.ID,
				})
			}
			break
		}

		// Handle incoming messages from client
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		logger.LogError("Failed to unmarshal client message", err, map[string]interface{}{
			"user_id":   c.UserID,
			"client_id": c.ID,
			"message":   string(message),
		})
		return
	}

	// Handle different message types
	switch msg.Type {
	case "ping":
		// Respond with pong
		response := Message{
			Type:      "pong",
			Timestamp: getCurrentTimestamp(),
		}
		c.sendMessage(response)

	case "subscribe":
		// Handle subscription to specific channels
		logger.Log.Info().
			Str("user_id", c.UserID).
			Str("client_id", c.ID).
			Interface("data", msg.Data).
			Msg("Client subscription request")

	case "heartbeat":
		// Client heartbeat - just log it
		logger.Log.Debug().
			Str("user_id", c.UserID).
			Str("client_id", c.ID).
			Msg("Client heartbeat received")

	default:
		logger.Log.Warn().
			Str("user_id", c.UserID).
			Str("client_id", c.ID).
			Str("message_type", msg.Type).
			Msg("Unknown message type from client")
	}
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.LogError("Failed to marshal message", err, map[string]interface{}{
			"user_id":   c.UserID,
			"client_id": c.ID,
		})
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
	}
}

// ServeWS handles websocket requests from the peer
func ServeWS(hub *Hub, c *gin.Context) {
	// Authenticate user via JWT token
	token := c.Query("token")
	if token == "" {
		// Try to get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No authentication token provided"})
		return
	}

	// Validate JWT token
	claims, err := auth.ValidateJWT(token)
	if err != nil {
		logger.LogSecurity("websocket_auth_failed", "", c.ClientIP(), map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.LogError("Failed to upgrade WebSocket connection", err, map[string]interface{}{
			"user_id": claims.UserID,
		})
		return
	}

	// Create client
	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		hub:      hub,
		UserID:   claims.UserID,
		ID:       generateClientID(),
		Username: claims.Username,
		Role:     claims.Role,
	}

	// Register client with hub
	client.hub.register <- client

	// Log successful connection
	logger.LogAudit("websocket_connected", claims.UserID, "", map[string]interface{}{
		"client_id": client.ID,
		"username":  claims.Username,
		"role":      claims.Role,
		"ip":        c.ClientIP(),
	})

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// generateClientID generates a unique client ID
func generateClientID() string {
	// Simple client ID generation - in production, use UUID
	return "client_" + string(rune(time.Now().UnixNano()%1000000))
}