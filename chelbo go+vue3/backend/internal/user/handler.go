package user

import (
	"encoding/json"
	"net/http"
	"strconv"

	"chelbo/backend/internal/auth"
	"chelbo/backend/internal/models"
	"chelbo/backend/internal/pkg/logger"
	"chelbo/backend/internal/pkg/validator"

	"github.com/jmoiron/sqlx"
)

type Handler struct {
	repo *Repository
	db   *sqlx.DB
}

func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{
		repo: NewRepository(db),
		db:   db,
	}
}

// GetProfile returns the current user's profile
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.repo.GetByID(userID)
	if err != nil {
		logger.Errorf("Failed to get user: %v", err)
		sendError(w, http.StatusNotFound, "user not found")
		return
	}

	sendSuccess(w, http.StatusOK, user.ToResponse())
}

// UpdateProfile updates the current user's profile
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate name if provided
	if req.Name != nil && !validator.IsValidName(*req.Name) {
		sendError(w, http.StatusBadRequest, "name must be between 2 and 100 characters")
		return
	}

	// Validate bio if provided
	if req.Bio != nil && !validator.IsValidBio(*req.Bio) {
		sendError(w, http.StatusBadRequest, "bio must be at most 200 characters")
		return
	}

	err := h.repo.UpdateProfile(userID, req.Name, req.Bio)
	if err != nil {
		logger.Errorf("Failed to update profile: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	// Get updated user
	user, err := h.repo.GetByID(userID)
	if err != nil {
		sendSuccess(w, http.StatusOK, map[string]string{"message": "profile updated"})
		return
	}

	sendSuccess(w, http.StatusOK, user.ToResponse())
}

// SearchUsers searches for users by name or email
func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" || len(query) < 2 {
		sendSuccess(w, http.StatusOK, []models.UserResponse{})
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	users, err := h.repo.SearchUsers(query, limit, userID)
	if err != nil {
		logger.Errorf("Failed to search users: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to search users")
		return
	}

	sendSuccess(w, http.StatusOK, users)
}

// GetUserByID returns a user by ID
func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 1 {
		sendError(w, http.StatusBadRequest, "invalid user ID")
		return
	}
	targetID, err := strconv.ParseUint(pathParts[len(pathParts)-1], 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.repo.GetByID(targetID)
	if err != nil {
		sendError(w, http.StatusNotFound, "user not found")
		return
	}

	sendSuccess(w, http.StatusOK, user.ToResponse())
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