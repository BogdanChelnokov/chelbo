package models

import (
	"time"
)

type ChatType string

const (
	ChatTypePrivate ChatType = "private"
	ChatTypeGroup   ChatType = "group"
)

type ParticipantRole string

const (
	RoleMember ParticipantRole = "member"
	RoleAdmin  ParticipantRole = "admin"
	RoleOwner  ParticipantRole = "owner"
)

type Chat struct {
	ID        uint64    `db:"id" json:"id"`
	Type      ChatType  `db:"type" json:"type"`
	Title     *string   `db:"title" json:"title,omitempty"`
	CreatedBy uint64    `db:"created_by" json:"created_by"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type ChatParticipant struct {
	ChatID     uint64          `db:"chat_id" json:"chat_id"`
	UserID     uint64          `db:"user_id" json:"user_id"`
	Role       ParticipantRole `db:"role" json:"role"`
	JoinedAt   time.Time       `db:"joined_at" json:"joined_at"`
	LastReadAt time.Time       `db:"last_read_at" json:"last_read_at"`
}

type ChatResponse struct {
	ID          uint64           `db:"id" json:"id"`
	Type        ChatType         `db:"type" json:"type"`
	Title       string           `db:"title" json:"title"`
	LastMessage *MessageResponse `json:"last_message,omitempty"`
	UnreadCount int              `db:"unread_count" json:"unread_count"`
	UpdatedAt   time.Time        `db:"updated_at" json:"updated_at"`
}

type CreateGroupRequest struct {
	Title             string   `json:"title" validate:"required,min=1,max=100"`
	ParticipantEmails []string `json:"participant_emails" validate:"required,min=1"`
}

type AddParticipantsRequest struct {
	Emails []string `json:"emails" validate:"required,min=1"`
}
