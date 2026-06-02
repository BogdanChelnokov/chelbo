package models

import (
	"time"
)

type User struct {
	ID           uint64    `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Name         string    `db:"name" json:"name"`
	AvatarURL    *string   `db:"avatar_url" json:"avatar_url,omitempty"`
	Bio          string    `db:"bio" json:"bio"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	IsVerified   bool      `db:"is_verified" json:"is_verified"`
	LastSeen     time.Time `db:"last_seen" json:"last_seen"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type UserResponse struct {
	ID       uint64    `db:"id" json:"id"`
	Email    string    `db:"email" json:"email"`
	Name     string    `db:"name" json:"name"`
	Bio      string    `db:"bio" json:"bio"`
	LastSeen time.Time `db:"last_seen" json:"last_seen"`
	IsOnline bool      `json:"is_online"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
	Name *string `json:"name,omitempty"`
	Bio  *string `json:"bio,omitempty"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:       u.ID,
		Email:    u.Email,
		Name:     u.Name,
		Bio:      u.Bio,
		LastSeen: u.LastSeen,
		IsOnline: false,
	}
}
