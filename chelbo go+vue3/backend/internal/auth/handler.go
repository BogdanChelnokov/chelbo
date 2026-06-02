package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"chelbo/backend/internal/models"
	"chelbo/backend/internal/pkg/config"
	"chelbo/backend/internal/pkg/logger"
	"chelbo/backend/internal/pkg/password"
	"chelbo/backend/internal/pkg/validator"

	"github.com/jmoiron/sqlx"
)

type AuthHandler struct {
	cfg *config.Config
	db  *sqlx.DB
}

func NewAuthHandler(cfg *config.Config, db *sqlx.DB) *AuthHandler {
	return &AuthHandler{
		cfg: cfg,
		db:  db,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if !validator.IsValidEmail(req.Email) {
		sendError(w, http.StatusBadRequest, "invalid email format")
		return
	}

	if !validator.IsValidPassword(req.Password) {
		sendError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	if !validator.IsValidName(req.Name) {
		sendError(w, http.StatusBadRequest, "name must be between 2 and 100 characters")
		return
	}

	// Check if user already exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)"
	err := h.db.Get(&exists, query, req.Email)
	if err != nil {
		logger.Errorf("Failed to check user existence: %v", err)
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if exists {
		sendError(w, http.StatusConflict, "user with this email already exists")
		return
	}

	// Hash password
	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		logger.Errorf("Failed to hash password: %v", err)
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Create user
	insertQuery := `
		INSERT INTO users (email, password_hash, name, bio)
		VALUES (?, ?, ?, '')
	`
	result, err := h.db.Exec(insertQuery, req.Email, hashedPassword, req.Name)
	if err != nil {
		logger.Errorf("Failed to create user: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Get the inserted ID
	userID, err := result.LastInsertId()
	if err != nil {
		logger.Errorf("Failed to get last insert id: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Get the created user - убираем avatar_url
	var user models.User
	selectQuery := `
		SELECT id, email, name, bio, is_active, is_verified, last_seen, created_at, updated_at
		FROM users 
		WHERE id = ?
	`
	err = h.db.Get(&user, selectQuery, userID)
	if err != nil {
		logger.Errorf("Failed to get created user: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(user.ID, user.Email, &h.cfg.JWT)
	if err != nil {
		logger.Errorf("Failed to generate token: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	logger.Infof("User registered successfully: %s (%d)", user.Email, user.ID)

	response := models.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}
	sendSuccess(w, http.StatusCreated, response)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if !validator.IsValidEmail(req.Email) {
		sendError(w, http.StatusBadRequest, "invalid email format")
		return
	}

	// Get user from database - убираем avatar_url
	var user models.User
	query := "SELECT id, email, password_hash, name, bio, is_active, is_verified, last_seen, created_at, updated_at FROM users WHERE email = ? AND is_active = TRUE"
	err := h.db.Get(&user, query, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			sendError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		logger.Errorf("Failed to get user: %v", err)
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Verify password
	if !password.Verify(user.PasswordHash, req.Password) {
		sendError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// Update last_seen
	_, err = h.db.Exec("UPDATE users SET last_seen = ? WHERE id = ?", time.Now(), user.ID)
	if err != nil {
		logger.Warnf("Failed to update last_seen: %v", err)
	}

	// Generate JWT token
	token, err := GenerateJWT(user.ID, user.Email, &h.cfg.JWT)
	if err != nil {
		logger.Errorf("Failed to generate token: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	logger.Infof("User logged in: %s (%d)", user.Email, user.ID)

	response := models.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}
	sendSuccess(w, http.StatusOK, response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if ok {
		logger.Infof("User logged out: %d", userID)
	}
	sendSuccess(w, http.StatusOK, models.NewMessageResponse("successfully logged out"))
}

// GetMe returns current user info
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "user not found in context")
		return
	}

	var user models.User
	query := "SELECT id, email, name, bio, avatar_url, is_active, is_verified, last_seen, created_at, updated_at FROM users WHERE id = ?"
	err := h.db.Get(&user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			sendError(w, http.StatusNotFound, "user not found")
			return
		}
		logger.Errorf("Failed to get user: %v", err)
		sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	sendSuccess(w, http.StatusOK, user.ToResponse())
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
