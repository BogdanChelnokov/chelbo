package models

import (
	"time"
)

type Message struct {
	ID            uint64    `db:"id" json:"id"`
	ChatID        uint64    `db:"chat_id" json:"chat_id"`
	SenderID      uint64    `db:"sender_id" json:"sender_id"`
	Text          *string   `db:"text" json:"text,omitempty"`
	FileURL       *string   `db:"file_url" json:"file_url,omitempty"`
	FileType      *string   `db:"file_type" json:"file_type,omitempty"`
	ReplyToID     *uint64   `db:"reply_to_id" json:"reply_to_id,omitempty"`
	ForwardedFrom *uint64   `db:"forwarded_from" json:"forwarded_from,omitempty"`
	IsDeleted     bool      `db:"is_deleted" json:"is_deleted"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type MessageResponse struct {
	ID            uint64           `db:"id" json:"id"`
	ChatID        uint64           `db:"chat_id" json:"chat_id"`
	SenderID      uint64           `db:"sender_id" json:"sender_id"`
	Text          *string          `db:"text" json:"text,omitempty"`
	FileURL       *string          `db:"file_url" json:"file_url,omitempty"`
	FileType      *string          `db:"file_type" json:"file_type,omitempty"`
	ReplyToID     *uint64          `db:"reply_to_id" json:"reply_to_id,omitempty"`
	ReplyTo       *MessageResponse `json:"reply_to,omitempty"`
	ForwardedFrom *uint64          `db:"forwarded_from" json:"forwarded_from,omitempty"`
	IsDelivered   bool             `db:"is_delivered" json:"is_delivered"`
	IsRead        bool             `db:"is_read" json:"is_read"`
	CreatedAt     time.Time        `db:"created_at" json:"created_at"`
}

type MessageStatus struct {
	MessageID   uint64     `db:"message_id" json:"message_id"`
	UserID      uint64     `db:"user_id" json:"user_id"`
	IsDelivered bool       `db:"is_delivered" json:"is_delivered"`
	IsRead      bool       `db:"is_read" json:"is_read"`
	DeliveredAt *time.Time `db:"delivered_at" json:"delivered_at,omitempty"`
	ReadAt      *time.Time `db:"read_at" json:"read_at,omitempty"`
}

type SendMessageRequest struct {
	Text      *string `json:"text,omitempty"`
	ReplyToID *uint64 `json:"reply_to_id,omitempty"`
}

type WebSocketMessage struct {
	Type      string           `json:"type"`
	ChatID    uint64           `json:"chat_id,omitempty"`
	MessageID uint64           `json:"message_id,omitempty"`
	Text      *string          `json:"text,omitempty"`
	IsTyping  bool             `json:"is_typing,omitempty"`
	Message   *MessageResponse `json:"message,omitempty"`
	UserID    uint64           `json:"user_id,omitempty"`
}
