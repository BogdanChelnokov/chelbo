package chat

import (
	"encoding/json"
	"net/http"
	"strconv"

	"chelbo/backend/internal/auth"
	"chelbo/backend/internal/models"
	"chelbo/backend/internal/pkg/logger"

	"github.com/jmoiron/sqlx"
)

type ChatHandler struct {
	chatRepo    *ChatRepository
	messageRepo *MessageRepository
	hub         *Hub
	db          *sqlx.DB
}

func NewChatHandler(db *sqlx.DB, hub *Hub) *ChatHandler {
	return &ChatHandler{
		chatRepo:    NewChatRepository(db),
		messageRepo: NewMessageRepository(db),
		hub:         hub,
		db:          db,
	}
}

// GetChats returns all chats for the current user
func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	chats, err := h.chatRepo.GetUserChats(userID)
	if err != nil {
		logger.Errorf("Failed to get chats: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to get chats")
		return
	}

	// For private chats, get the partner's name as title
	for i := range chats {
		if chats[i].Type == "private" {
			partner, err := h.chatRepo.GetPrivateChatPartner(chats[i].ID, userID)
			if err == nil && partner != nil {
				chats[i].Title = partner.Name
			}
		}
	}

	sendSuccess(w, http.StatusOK, chats)
}

// CreatePrivateChat creates a private chat with another user
func (h *ChatHandler) CreatePrivateChat(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get user ID from URL
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 1 {
		sendError(w, http.StatusBadRequest, "invalid user ID")
		return
	}
	otherUserID, err := strconv.ParseUint(pathParts[len(pathParts)-1], 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	chatID, err := h.chatRepo.CreatePrivateChat(userID, otherUserID)
	if err != nil {
		logger.Errorf("Failed to create private chat: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to create chat")
		return
	}

	sendSuccess(w, http.StatusCreated, map[string]uint64{"chat_id": chatID})
}

// CreateGroup creates a new group chat
func (h *ChatHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get user IDs from emails
	var participantIDs []uint64
	for _, email := range req.ParticipantEmails {
		var id uint64
		err := h.db.Get(&id, "SELECT id FROM users WHERE email = ?", email)
		if err != nil {
			continue
		}
		participantIDs = append(participantIDs, id)
	}
	participantIDs = append(participantIDs, userID)

	chatID, err := h.chatRepo.CreateGroup(userID, req.Title, participantIDs)
	if err != nil {
		logger.Errorf("Failed to create group: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to create group")
		return
	}

	sendSuccess(w, http.StatusCreated, map[string]uint64{"chat_id": chatID})
}

// GetMessages returns message history for a chat
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get chat ID from URL
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 2 {
		sendError(w, http.StatusBadRequest, "invalid chat ID")
		return
	}
	chatID, err := strconv.ParseUint(pathParts[len(pathParts)-2], 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid chat ID")
		return
	}

	// Check if user is participant
	isParticipant, err := h.chatRepo.IsParticipant(chatID, userID)
	if err != nil || !isParticipant {
		sendError(w, http.StatusForbidden, "access denied")
		return
	}

	// Get pagination parameters
	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	messages, err := h.messageRepo.GetMessages(chatID, limit, offset, userID)
	if err != nil {
		logger.Errorf("Failed to get messages: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to get messages")
		return
	}

	sendSuccess(w, http.StatusOK, messages)
}

// SendMessage sends a new message
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get chat ID from URL
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 2 {
		sendError(w, http.StatusBadRequest, "invalid chat ID")
		return
	}
	chatID, err := strconv.ParseUint(pathParts[len(pathParts)-2], 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid chat ID")
		return
	}

	// Check if user is participant
	isParticipant, err := h.chatRepo.IsParticipant(chatID, userID)
	if err != nil || !isParticipant {
		sendError(w, http.StatusForbidden, "access denied")
		return
	}

	var req struct {
		Text      *string `json:"text"`
		FileURL   *string `json:"file_url"`
		FileType  *string `json:"file_type"`
		ReplyToID *uint64 `json:"reply_to_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Failed to decode request: %v", err)
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Проверяем, есть ли текст или файл
	if (req.Text == nil || *req.Text == "") && (req.FileURL == nil || *req.FileURL == "") {
		sendError(w, http.StatusBadRequest, "message text or file is required")
		return
	}

	// Create message - 7 аргументов (forwardedFrom = nil для обычных сообщений)
	message, err := h.messageRepo.CreateMessage(
		chatID, 
		userID, 
		req.Text, 
		req.FileURL, 
		req.FileType, 
		req.ReplyToID,
		nil, // forwardedFrom
	)
	if err != nil {
		logger.Errorf("Failed to create message: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to send message")
		return
	}

	// Create response
	msgResponse := models.MessageResponse{
		ID:          message.ID,
		ChatID:      message.ChatID,
		SenderID:    message.SenderID,
		Text:        message.Text,
		FileURL:     message.FileURL,
		FileType:    message.FileType,
		IsDelivered: false,
		IsRead:      false,
		CreatedAt:   message.CreatedAt,
	}

	// Broadcast via WebSocket
	wsMessage := models.WebSocketMessage{
		Type:    "new_message",
		Message: &msgResponse,
		ChatID:  chatID,
	}
	h.hub.SendToChat(chatID, wsMessage, userID)

	sendSuccess(w, http.StatusCreated, msgResponse)
}

// DeleteMessage deletes a message
func (h *ChatHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get message ID from URL
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 1 {
		sendError(w, http.StatusBadRequest, "invalid message ID")
		return
	}
	messageID, err := strconv.ParseUint(pathParts[len(pathParts)-1], 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid message ID")
		return
	}

	err = h.messageRepo.DeleteMessage(messageID, userID)
	if err != nil {
		logger.Errorf("Failed to delete message: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to delete message")
		return
	}

	sendSuccess(w, http.StatusOK, map[string]string{"message": "message deleted"})
}

func splitPath(path string) []string {
	parts := make([]string, 0)
	start := 1
	for i := 1; i < len(path); i++ {
		if path[i] == '/' {
			if start < i {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}

func sendSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.NewSuccessResponse(data))
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.NewErrorResponse(message))
}

// MarkAsRead marks a message as read for the current user
func (h *ChatHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get message ID from URL
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 1 {
		sendError(w, http.StatusBadRequest, "invalid message ID")
		return
	}
	messageID, err := strconv.ParseUint(pathParts[len(pathParts)-2], 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid message ID")
		return
	}

	// Mark message as read
	err = h.messageRepo.MarkAsRead(messageID, userID)
	if err != nil {
		logger.Errorf("Failed to mark message as read: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to mark message as read")
		return
	}

	sendSuccess(w, http.StatusOK, map[string]string{"message": "marked as read"})
}

// ForwardMessage forwards a message to another chat
func (h *ChatHandler) ForwardMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get message ID from URL
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 2 {
		sendError(w, http.StatusBadRequest, "invalid message ID")
		return
	}
	messageID, err := strconv.ParseUint(pathParts[len(pathParts)-2], 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid message ID")
		return
	}

	var req struct {
		TargetChatID uint64 `json:"target_chat_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get original message - включаем sender_id
	var originalMessage struct {
		Text     *string `db:"text"`
		FileURL  *string `db:"file_url"`
		FileType *string `db:"file_type"`
		SenderID uint64  `db:"sender_id"`
	}
	err = h.db.Get(&originalMessage, "SELECT text, file_url, file_type, sender_id FROM messages WHERE id = ?", messageID)
	if err != nil {
		logger.Errorf("Failed to get original message: %v", err)
		sendError(w, http.StatusNotFound, "message not found")
		return
	}

	// Create new message with forwarded_from
	newMessage, err := h.messageRepo.CreateMessage(
		req.TargetChatID,
		userID,
		originalMessage.Text,
		originalMessage.FileURL,
		originalMessage.FileType,
		nil,                       // replyToID
		&originalMessage.SenderID, // forwarded_from - ID оригинального отправителя
	)
	if err != nil {
		logger.Errorf("Failed to forward message: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to forward message")
		return
	}

	sendSuccess(w, http.StatusCreated, newMessage)
}
