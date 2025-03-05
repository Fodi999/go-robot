package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message represents a chat message with a read flag.
type Message struct {
	ChatID    string `json:"chat_id"`
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
	Read      bool   `json:"read"`
}

// StatusMessage represents the chat status with online users.
type StatusMessage struct {
	Type        string   `json:"type"`
	OnlineUsers []string `json:"onlineUsers"`
}

// ChatState represents the current state of a chat.
type ChatState struct {
	Messages    []Message `json:"messages"`
	OnlineUsers []string  `json:"onlineUsers"`
}

// ReadStatusUpdate represents an update to a message's read status.
type ReadStatusUpdate struct {
	Update    string `json:"update"` // Expected value: "read"
	ChatID    string `json:"chat_id"`
	Timestamp int64  `json:"timestamp"`
	Read      bool   `json:"read"`
}

// ChatHub manages chats and message broadcasting.
type ChatHub struct {
	chats     map[string]map[*websocket.Conn]string // Stores username instead of just client_id.
	messages  map[string][]Message                  // Message history by chatID.
	broadcast chan Message                          // Channel for new messages.
	mu        sync.Mutex                            // Synchronization for concurrent access.
}

// NewChatHub creates a new ChatHub instance.
func NewChatHub() *ChatHub {
	return &ChatHub{
		chats:     make(map[string]map[*websocket.Conn]string),
		messages:  make(map[string][]Message),
		broadcast: make(chan Message),
	}
}

// Run processes incoming messages and broadcasts them to all clients.
func (hub *ChatHub) Run() {
	for msg := range hub.broadcast {
		hub.mu.Lock()
		clients, exists := hub.chats[msg.ChatID]
		if !exists {
			hub.mu.Unlock()
			continue
		}

		// Set timestamp if not provided.
		if msg.Timestamp == 0 {
			msg.Timestamp = time.Now().UnixMilli()
		}
		// Mark message as read if there are 2 or more clients (e.g., client and admin).
		if len(clients) >= 2 {
			msg.Read = true
		} else {
			msg.Read = false
		}
		// Save message to history.
		hub.messages[msg.ChatID] = append(hub.messages[msg.ChatID], msg)

		// Send message to all clients.
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Error marshaling message: %v", err)
			hub.mu.Unlock()
			continue
		}
		for client := range clients {
			if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Error sending to client in chat %s: %v", msg.ChatID, err)
				client.Close()
				delete(clients, client)
			}
		}

		hub.broadcastStatus(msg.ChatID)
		hub.mu.Unlock()
	}
}

// broadcastStatus sends the status (list of online users) to the specified chat.
func (hub *ChatHub) broadcastStatus(chatID string) {
	clients, exists := hub.chats[chatID]
	if !exists {
		return
	}
	onlineUsers := make([]string, 0, len(clients))
	for _, username := range clients {
		onlineUsers = append(onlineUsers, username)
	}
	statusMsg := StatusMessage{
		Type:        "status",
		OnlineUsers: onlineUsers,
	}
	data, err := json.Marshal(statusMsg)
	if err != nil {
		log.Printf("Error marshaling status: %v", err)
		return
	}
	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error sending status in chat %s: %v", chatID, err)
			client.Close()
			delete(clients, client)
		}
	}
}

// sendChatState sends the current chat state (messages and online users) to a new client.
func (hub *ChatHub) sendChatState(conn *websocket.Conn, chatID string) {
	clients := hub.chats[chatID]
	onlineUsers := make([]string, 0, len(clients))
	for _, username := range clients {
		onlineUsers = append(onlineUsers, username)
	}
	state := ChatState{
		Messages:    hub.messages[chatID],
		OnlineUsers: onlineUsers,
	}
	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("Error marshaling chat state %s: %v", chatID, err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Error sending chat state %s: %v", chatID, err)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Add Origin check for production.
}

// ChatHandler handles WebSocket connections and message exchange.
func (hub *ChatHub) ChatHandler(w http.ResponseWriter, r *http.Request) {
	// Extract client_id and username.
	clientID := r.URL.Query().Get("client_id")
	username := r.URL.Query().Get("username")
	if clientID == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}
	// Use clientID as username if not provided.
	if username == "" {
		username = clientID
	}
	// chat_id is optional at connection time.
	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		chatID = "global" // Use a default chatID if not specified.
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	hub.mu.Lock()
	if _, exists := hub.chats[chatID]; !exists {
		hub.chats[chatID] = make(map[*websocket.Conn]string)
		hub.messages[chatID] = []Message{}
	}
	hub.chats[chatID][ws] = username
	hub.mu.Unlock()

	log.Printf("Client %s (%s) connected to chat %s", clientID, username, chatID)
	hub.sendChatState(ws, chatID)
	hub.broadcastStatus(chatID)

	go hub.pingLoop(ws, clientID, chatID)

	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Client %s disconnected from chat %s: %v", clientID, chatID, err)
			hub.mu.Lock()
			delete(hub.chats[chatID], ws)
			if len(hub.chats[chatID]) == 0 {
				delete(hub.chats, chatID)
				delete(hub.messages, chatID)
			}
			hub.broadcastStatus(chatID)
			hub.mu.Unlock()
			break
		}

		// Handle read status updates.
		var potentialUpdate struct {
			Update string `json:"update"`
		}
		if err := json.Unmarshal(data, &potentialUpdate); err == nil && potentialUpdate.Update == "read" {
			var update ReadStatusUpdate
			if err := json.Unmarshal(data, &update); err != nil {
				log.Printf("Error parsing read status update: %v", err)
				continue
			}
			hub.mu.Lock()
			msgs := hub.messages[update.ChatID]
			for i, m := range msgs {
				if m.Timestamp == update.Timestamp {
					hub.messages[update.ChatID][i].Read = update.Read
					updatedMsg, err := json.Marshal(hub.messages[update.ChatID][i])
					if err != nil {
						log.Printf("Error marshaling updated message: %v", err)
						continue
					}
					for client := range hub.chats[update.ChatID] {
						if err := client.WriteMessage(websocket.TextMessage, updatedMsg); err != nil {
							log.Printf("Error sending updated status: %v", err)
							client.Close()
							delete(hub.chats[update.ChatID], client)
						}
					}
				}
			}
			hub.mu.Unlock()
			continue
		}

		// Handle text messages.
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Error parsing JSON from %s: %v", clientID, err)
			continue
		}

		// Ensure chat_id is provided in the message.
		if msg.ChatID == "" {
			log.Printf("Error: chat_id not provided in message from %s", clientID)
			// Send error message to the client.
			errorMsg := map[string]string{"error": "chat_id is required"}
			data, _ := json.Marshal(errorMsg)
			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Error sending error to client %s: %v", clientID, err)
			}
			continue
		}

		// Verify client exists in the chat and set sender.
		hub.mu.Lock()
		senderName, exists := hub.chats[chatID][ws]
		if !exists {
			log.Printf("Error: client %s not found in chat %s", clientID, chatID)
			hub.mu.Unlock()
			continue
		}
		hub.mu.Unlock()
		msg.Sender = senderName
		if msg.Timestamp == 0 {
			msg.Timestamp = time.Now().UnixMilli()
		}

		log.Printf("Received from %s in chat %s: %s", clientID, msg.ChatID, msg.Text)
		hub.broadcast <- msg
	}
}

// pingLoop sends ping messages to keep the connection alive.
func (hub *ChatHub) pingLoop(conn *websocket.Conn, clientID, chatID string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		hub.mu.Lock()
		_, exists := hub.chats[chatID][conn]
		hub.mu.Unlock()
		if !exists {
			return
		}
		if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
			log.Printf("Error sending ping to client %s in chat %s: %v", clientID, chatID, err)
			return
		}
	}
}















