package chat

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"chelbo/backend/internal/models"
	"chelbo/backend/internal/pkg/logger"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client represents a WebSocket client
type Client struct {
	ID     uint64
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
	UserID uint64
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	Clients    map[uint64]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	ChatID   uint64
	Message  interface{}
	SenderID uint64
}

var hubInstance *Hub
var hubOnce sync.Once

// GetHub returns the singleton hub instance
func GetHub() *Hub {
	hubOnce.Do(func() {
		hubInstance = &Hub{
			Clients:    make(map[uint64]*Client),
			Register:   make(chan *Client),
			Unregister: make(chan *Client),
			Broadcast:  make(chan *BroadcastMessage, 100),
		}
		go hubInstance.Run()
	})
	return hubInstance
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.UserID] = client
			h.mu.Unlock()
			logger.Infof("Client registered: %d", client.UserID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				close(client.Send)
				logger.Infof("Client unregistered: %d", client.UserID)
			}
			h.mu.Unlock()

		case broadcast := <-h.Broadcast:
			h.mu.RLock()
			for _, client := range h.Clients {
				if client.UserID == broadcast.SenderID {
					continue
				}
				select {
				case client.Send <- encodeMessage(broadcast.Message):
				default:
					close(client.Send)
					delete(h.Clients, client.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID uint64, message interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.Clients[userID]; ok {
		select {
		case client.Send <- encodeMessage(message):
		default:
			close(client.Send)
			delete(h.Clients, userID)
		}
	}
}

// SendToChat sends a message to all users in a chat (except sender)
func (h *Hub) SendToChat(chatID uint64, message interface{}, senderID uint64) {
	logger.Infof("Broadcasting to chat %d from user %d", chatID, senderID)
	h.Broadcast <- &BroadcastMessage{
		ChatID:   chatID,
		Message:  message,
		SenderID: senderID,
	}
}

func encodeMessage(msg interface{}) []byte {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Errorf("Failed to encode message: %v", err)
		return nil
	}
	return data
}

// HandleWebSocket handles WebSocket connections
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID uint64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("Failed to upgrade to WebSocket: %v", err)
		return
	}

	client := &Client{
		ID:     userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h,
		UserID: userID,
	}

	h.Register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg models.WebSocketMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		switch msg.Type {
		case "message":
			logger.Debugf("Received message from user %d in chat %d", c.UserID, msg.ChatID)
		case "typing":
			typingMsg := models.WebSocketMessage{
				Type:     "typing",
				ChatID:   msg.ChatID,
				UserID:   c.UserID,
				IsTyping: msg.IsTyping,
			}
			c.Hub.SendToChat(msg.ChatID, typingMsg, c.UserID)
		case "read":
			logger.Debugf("User %d read message %d", c.UserID, msg.MessageID)
		}
	}
}
