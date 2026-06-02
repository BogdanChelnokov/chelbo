package chat

import (
	"database/sql"
	"time"

	"chelbo/backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type ChatRepository struct {
	db *sqlx.DB
}

func NewChatRepository(db *sqlx.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// CreatePrivateChat creates a private chat between two users
func (r *ChatRepository) CreatePrivateChat(user1ID, user2ID uint64) (uint64, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Check if private chat already exists
	var existingChatID uint64
	query := `
		SELECT c.id FROM chats c
		INNER JOIN chat_participants cp1 ON c.id = cp1.chat_id
		INNER JOIN chat_participants cp2 ON c.id = cp2.chat_id
		WHERE c.type = 'private' 
		AND cp1.user_id = ? 
		AND cp2.user_id = ?
	`
	err = tx.Get(&existingChatID, query, user1ID, user2ID)
	if err == nil {
		return existingChatID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	// Create new private chat
	insertChat := `
		INSERT INTO chats (type, created_by) 
		VALUES ('private', ?)
	`
	result, err := tx.Exec(insertChat, user1ID)
	if err != nil {
		return 0, err
	}

	chatID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Add participants
	insertParticipant := `
		INSERT INTO chat_participants (chat_id, user_id, role)
		VALUES (?, ?, ?)
	`
	_, err = tx.Exec(insertParticipant, chatID, user1ID, "owner")
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(insertParticipant, chatID, user2ID, "member")
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return uint64(chatID), nil
}

// CreateGroup creates a new group chat
func (r *ChatRepository) CreateGroup(createdBy uint64, title string, participantIDs []uint64) (uint64, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	insertChat := `
		INSERT INTO chats (type, title, created_by) 
		VALUES ('group', ?, ?)
	`
	result, err := tx.Exec(insertChat, title, createdBy)
	if err != nil {
		return 0, err
	}

	chatID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Add creator as owner
	insertParticipant := `
		INSERT INTO chat_participants (chat_id, user_id, role)
		VALUES (?, ?, ?)
	`
	_, err = tx.Exec(insertParticipant, chatID, createdBy, "owner")
	if err != nil {
		return 0, err
	}

	// Add other participants as members
	for _, userID := range participantIDs {
		if userID == createdBy {
			continue
		}
		_, err = tx.Exec(insertParticipant, chatID, userID, "member")
		if err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return uint64(chatID), nil
}

// GetUserChats returns all chats for a user with unread count and last message
func (r *ChatRepository) GetUserChats(userID uint64) ([]models.ChatResponse, error) {
	query := `
		SELECT 
			c.id, 
			c.type, 
			COALESCE(c.title, '') as title,
			c.updated_at,
			COALESCE(
				(SELECT COUNT(*) 
				 FROM messages m 
				 INNER JOIN message_statuses ms ON m.id = ms.message_id 
				 WHERE m.chat_id = c.id 
				 AND ms.user_id = ? 
				 AND ms.is_read = FALSE 
				 AND m.sender_id != ?), 0
			) as unread_count
		FROM chats c
		INNER JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = ?
		ORDER BY c.updated_at DESC
	`

	var chats []models.ChatResponse
	err := r.db.Select(&chats, query, userID, userID, userID)
	if err != nil {
		return nil, err
	}

	// Для каждого чата получаем последнее сообщение отдельно
	for i, chat := range chats {
		var lastMessage models.MessageResponse
		msgQuery := `
			SELECT 
				m.id, 
				m.chat_id, 
				m.sender_id, 
				COALESCE(m.text, '') as text,
				m.created_at,
				COALESCE(ms.is_delivered, FALSE) as is_delivered,
				COALESCE(ms.is_read, FALSE) as is_read
			FROM messages m
			LEFT JOIN message_statuses ms ON m.id = ms.message_id AND ms.user_id = ?
			WHERE m.chat_id = ? AND m.is_deleted = FALSE
			ORDER BY m.created_at DESC
			LIMIT 1
		`
		err = r.db.Get(&lastMessage, msgQuery, userID, chat.ID)
		if err == nil {
			chats[i].LastMessage = &lastMessage
		}
	}

	if chats == nil {
		return []models.ChatResponse{}, nil
	}

	return chats, nil
}

// GetChatParticipants returns all participants of a chat
func (r *ChatRepository) GetChatParticipants(chatID uint64) ([]models.UserResponse, error) {
	query := `
		SELECT u.id, u.email, u.name, u.bio, u.last_seen
		FROM users u
		INNER JOIN chat_participants cp ON u.id = cp.user_id
		WHERE cp.chat_id = ?
	`
	var users []models.UserResponse
	err := r.db.Select(&users, query, chatID)
	return users, err
}

// IsParticipant checks if a user is a participant of a chat
func (r *ChatRepository) IsParticipant(chatID, userID uint64) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM chat_participants WHERE chat_id = ? AND user_id = ?"
	err := r.db.Get(&count, query, chatID, userID)
	return count > 0, err
}

// AddParticipants adds users to a group chat
func (r *ChatRepository) AddParticipants(chatID uint64, userIDs []uint64) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertParticipant := `
		INSERT IGNORE INTO chat_participants (chat_id, user_id, role)
		VALUES (?, ?, 'member')
	`
	for _, userID := range userIDs {
		_, err = tx.Exec(insertParticipant, chatID, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RemoveParticipant removes a user from a group chat
func (r *ChatRepository) RemoveParticipant(chatID, userID uint64) error {
	query := "DELETE FROM chat_participants WHERE chat_id = ? AND user_id = ?"
	_, err := r.db.Exec(query, chatID, userID)
	return err
}

// UpdateLastRead updates the last_read_at timestamp for a user in a chat
func (r *ChatRepository) UpdateLastRead(chatID, userID uint64) error {
	query := "UPDATE chat_participants SET last_read_at = ? WHERE chat_id = ? AND user_id = ?"
	_, err := r.db.Exec(query, time.Now(), chatID, userID)
	return err
}

// GetPrivateChatPartner returns the other user in a private chat
func (r *ChatRepository) GetPrivateChatPartner(chatID, userID uint64) (*models.UserResponse, error) {
	query := `
		SELECT u.id, u.email, u.name, u.bio, u.last_seen
		FROM users u
		INNER JOIN chat_participants cp ON u.id = cp.user_id
		WHERE cp.chat_id = ? AND cp.user_id != ?
		LIMIT 1
	`
	var user models.UserResponse
	err := r.db.Get(&user, query, chatID, userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
