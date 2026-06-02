package chat

import (
	"database/sql"
	"time"

	"chelbo/backend/internal/models"
	"chelbo/backend/internal/pkg/logger"

	"github.com/jmoiron/sqlx"
)

type MessageRepository struct {
	db *sqlx.DB
}

func NewMessageRepository(db *sqlx.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) CreateMessage(chatID, senderID uint64, text *string, fileURL, fileType *string, replyToID, forwardedFrom *uint64) (*models.Message, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert message - убедись что forwarded_from включен в запрос
	insertQuery := `
		INSERT INTO messages (chat_id, sender_id, text, file_url, file_type, reply_to_id, forwarded_from)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := tx.Exec(insertQuery, chatID, senderID, text, fileURL, fileType, replyToID, forwardedFrom)
	if err != nil {
		logger.Errorf("Failed to insert message: %v", err)
		return nil, err
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Get the created message
	var message models.Message
	selectQuery := `
		SELECT id, chat_id, sender_id, text, file_url, file_type, reply_to_id, forwarded_from, is_deleted, created_at
		FROM messages 
		WHERE id = ?
	`
	err = tx.Get(&message, selectQuery, messageID)
	if err != nil {
		return nil, err
	}

	// Get all participants except sender
	var participants []uint64
	participantQuery := `
		SELECT user_id FROM chat_participants 
		WHERE chat_id = ? AND user_id != ?
	`
	err = tx.Select(&participants, participantQuery, chatID, senderID)
	if err != nil {
		return nil, err
	}

	// Create message statuses for all participants
	insertStatus := `
		INSERT INTO message_statuses (message_id, user_id, is_delivered, is_read)
		VALUES (?, ?, FALSE, FALSE)
	`
	for _, userID := range participants {
		_, err = tx.Exec(insertStatus, messageID, userID)
		if err != nil {
			return nil, err
		}
	}

	// Update chat updated_at
	_, err = tx.Exec("UPDATE chats SET updated_at = ? WHERE id = ?", time.Now(), chatID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &message, nil
}

// GetMessages returns messages for a chat with pagination and reply_to data
func (r *MessageRepository) GetMessages(chatID uint64, limit, offset int, userID uint64) ([]models.MessageResponse, error) {
	query := `
		SELECT 
			m.id, 
			m.chat_id, 
			m.sender_id, 
			m.text, 
			m.file_url, 
			m.file_type, 
			m.reply_to_id,
			m.created_at,
			COALESCE(ms.is_delivered, FALSE) as is_delivered,
			COALESCE(ms.is_read, FALSE) as is_read
		FROM messages m
		LEFT JOIN message_statuses ms ON m.id = ms.message_id AND ms.user_id = ?
		WHERE m.chat_id = ? AND m.is_deleted = FALSE
		ORDER BY m.created_at ASC
		LIMIT ? OFFSET ?
	`

	var messages []models.MessageResponse
	err := r.db.Select(&messages, query, userID, chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Загружаем reply_to сообщения для каждого сообщения, у которого есть reply_to_id
	for i, msg := range messages {
		if msg.ReplyToID != nil && *msg.ReplyToID > 0 {
			var replyTo models.MessageResponse
			replyQuery := `
				SELECT id, chat_id, sender_id, text, file_url, file_type, created_at
				FROM messages 
				WHERE id = ? AND is_deleted = FALSE
			`
			err = r.db.Get(&replyTo, replyQuery, *msg.ReplyToID)
			if err == nil {
				messages[i].ReplyTo = &replyTo
			}
		}
	}

	return messages, nil
}

// MarkAsDelivered marks a message as delivered for a user
func (r *MessageRepository) MarkAsDelivered(messageID, userID uint64) error {
	query := `
		UPDATE message_statuses 
		SET is_delivered = TRUE, delivered_at = ?
		WHERE message_id = ? AND user_id = ? AND is_delivered = FALSE
	`
	_, err := r.db.Exec(query, time.Now(), messageID, userID)
	return err
}

// MarkAsRead marks a message as read for a user
func (r *MessageRepository) MarkAsRead(messageID, userID uint64) error {
	query := `
		UPDATE message_statuses 
		SET is_read = TRUE, read_at = ?
		WHERE message_id = ? AND user_id = ? AND is_read = FALSE
	`
	_, err := r.db.Exec(query, time.Now(), messageID, userID)
	return err
}

// GetUnreadMessages returns all unread messages for a user in a chat
func (r *MessageRepository) GetUnreadMessages(chatID, userID uint64) ([]models.MessageResponse, error) {
	query := `
		SELECT 
			m.id, m.chat_id, m.sender_id, m.text, m.file_url, m.file_type, m.created_at,
			ms.is_delivered, ms.is_read
		FROM messages m
		INNER JOIN message_statuses ms ON m.id = ms.message_id
		WHERE m.chat_id = ? AND ms.user_id = ? AND ms.is_read = FALSE AND m.sender_id != ?
		ORDER BY m.created_at ASC
	`
	var messages []models.MessageResponse
	err := r.db.Select(&messages, query, chatID, userID, userID)
	return messages, err
}

// DeleteMessage soft deletes a message
func (r *MessageRepository) DeleteMessage(messageID, userID uint64) error {
	// Check if user is the sender
	var senderID uint64
	query := "SELECT sender_id FROM messages WHERE id = ?"
	err := r.db.Get(&senderID, query, messageID)
	if err != nil {
		return err
	}

	if senderID != userID {
		return sql.ErrNoRows
	}

	_, err = r.db.Exec("UPDATE messages SET is_deleted = TRUE WHERE id = ?", messageID)
	return err
}
