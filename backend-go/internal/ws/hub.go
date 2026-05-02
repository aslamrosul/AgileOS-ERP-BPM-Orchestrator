package ws

import (
	"encoding/json"
	"sync"

	"agileos-backend/logger"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients mapped by user ID
	clients map[string]map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mutex sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	UserID    string                 `json:"user_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// NotificationMessage represents a notification to be sent to users
type NotificationMessage struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // task_assigned, task_completed, process_started, etc.
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	UserID      string                 `json:"user_id"`
	TaskID      string                 `json:"task_id,omitempty"`
	ProcessID   string                 `json:"process_id,omitempty"`
	WorkflowID  string                 `json:"workflow_id,omitempty"`
	Priority    string                 `json:"priority"` // low, medium, high, urgent
	ActionURL   string                 `json:"action_url,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   int64                  `json:"timestamp"`
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub and handles client registration/unregistration
func (h *Hub) Run() {
	logger.Log.Info().Msg("🔌 WebSocket Hub started")

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]bool)
	}
	h.clients[client.UserID][client] = true

	logger.Log.Info().
		Str("user_id", client.UserID).
		Str("client_id", client.ID).
		Int("total_connections", len(h.clients[client.UserID])).
		Msg("WebSocket client registered")

	// Send welcome message
	welcomeMsg := NotificationMessage{
		ID:        "welcome_" + client.ID,
		Type:      "connection_established",
		Title:     "Connected",
		Message:   "Real-time notifications are now active",
		UserID:    client.UserID,
		Priority:  "low",
		Timestamp: getCurrentTimestamp(),
	}

	h.SendToUser(client.UserID, welcomeMsg)
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if clients, ok := h.clients[client.UserID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.send)

			// Remove user entry if no more clients
			if len(clients) == 0 {
				delete(h.clients, client.UserID)
			}

			logger.Log.Info().
				Str("user_id", client.UserID).
				Str("client_id", client.ID).
				Msg("WebSocket client unregistered")
		}
	}
}

// broadcastMessage broadcasts a message to all connected clients
func (h *Hub) broadcastMessage(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for userID, clients := range h.clients {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				// Client's send channel is full, close it
				close(client.send)
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.clients, userID)
				}
			}
		}
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID string, notification NotificationMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients, exists := h.clients[userID]
	if !exists {
		logger.Log.Debug().
			Str("user_id", userID).
			Str("notification_type", notification.Type).
			Msg("User not connected, notification not sent")
		return
	}

	message, err := json.Marshal(notification)
	if err != nil {
		logger.LogError("Failed to marshal notification", err, map[string]interface{}{
			"user_id": userID,
			"type":    notification.Type,
		})
		return
	}

	// Send to all client connections for this user
	for client := range clients {
		select {
		case client.send <- message:
			logger.Log.Debug().
				Str("user_id", userID).
				Str("client_id", client.ID).
				Str("notification_type", notification.Type).
				Msg("Notification sent to client")
		default:
			// Client's send channel is full, close it
			close(client.send)
			delete(clients, client)
		}
	}

	// Clean up if no more clients
	if len(clients) == 0 {
		delete(h.clients, userID)
	}

	// Log notification delivery
	logger.LogAudit("notification_sent", userID, notification.TaskID, map[string]interface{}{
		"notification_type": notification.Type,
		"title":            notification.Title,
		"priority":         notification.Priority,
	})
}

// SendToMultipleUsers sends a notification to multiple users
func (h *Hub) SendToMultipleUsers(userIDs []string, notification NotificationMessage) {
	for _, userID := range userIDs {
		// Create a copy of notification for each user
		userNotification := notification
		userNotification.UserID = userID
		h.SendToUser(userID, userNotification)
	}
}

// BroadcastToAll sends a message to all connected users
func (h *Hub) BroadcastToAll(notification NotificationMessage) {
	h.mutex.RLock()
	userIDs := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		userIDs = append(userIDs, userID)
	}
	h.mutex.RUnlock()

	h.SendToMultipleUsers(userIDs, notification)
}

// GetConnectedUsers returns a list of currently connected user IDs
func (h *Hub) GetConnectedUsers() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	users := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// GetConnectionCount returns the total number of active connections
func (h *Hub) GetConnectionCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}

// IsUserConnected checks if a user is currently connected
func (h *Hub) IsUserConnected(userID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients, exists := h.clients[userID]
	return exists && len(clients) > 0
}

// Helper function to get current timestamp
func getCurrentTimestamp() int64 {
	return 1777378870 // Current timestamp
}