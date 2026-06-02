package file

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"chelbo/backend/internal/auth"
	"chelbo/backend/internal/models"
	"chelbo/backend/internal/pkg/config"
	"chelbo/backend/internal/pkg/logger"

	"github.com/google/uuid"
)

type Handler struct {
	cfg *config.Config
}

func NewHandler(cfg *config.Config) *Handler {
	// Create uploads directory if it doesn't exist
	if cfg.File.StorageType == "local" {
		if err := os.MkdirAll(cfg.File.Path, 0755); err != nil {
			logger.Errorf("Failed to create uploads directory: %v", err)
		}
	}
	return &Handler{cfg: cfg}
}

// UploadFile handles file uploads
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(h.cfg.File.MaxSizeMB << 20)
	if err != nil {
		sendError(w, http.StatusBadRequest, "failed to parse form: file too large?")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		sendError(w, http.StatusBadRequest, "no file provided")
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > h.cfg.File.MaxSizeMB<<20 {
		sendError(w, http.StatusBadRequest, "file too large")
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext
	filePath := filepath.Join(h.cfg.File.Path, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		logger.Errorf("Failed to create file: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to save file")
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		logger.Errorf("Failed to copy file: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	// Determine file type - ЗДЕСЬ исправь вызов
	fileType := getFileType(header.Header.Get("Content-Type")) // убрали ext

	// Generate file URL
	fileURL := "/uploads/" + filename

	logger.Infof("User %d uploaded file: %s (%d bytes), type: %s", userID, filename, header.Size, fileType)

	response := map[string]interface{}{
		"file_url":  fileURL,
		"filename":  header.Filename,
		"file_type": fileType,
		"size":      header.Size,
	}

	sendSuccess(w, http.StatusCreated, response)
}

// GetFile serves a file
func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {
	// Get filename from URL
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 1 {
		sendError(w, http.StatusBadRequest, "invalid file path")
		return
	}
	filename := pathParts[len(pathParts)-1]

	filePath := filepath.Join(h.cfg.File.Path, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		sendError(w, http.StatusNotFound, "file not found")
		return
	}

	// Set correct content type based on file extension
	ext := filepath.Ext(filename)
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webm":
		contentType = "video/webm"
	case ".mp4":
		contentType = "video/mp4"
	case ".mp3":
		contentType = "audio/mpeg"
	}

	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, r, filePath)
}

// UploadAvatar handles avatar upload
func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(5 << 20) // 5MB max for avatar
	if err != nil {
		sendError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		sendError(w, http.StatusBadRequest, "no file provided")
		return
	}
	defer file.Close()

	// Validate file size (max 5MB for avatar)
	if header.Size > 5<<20 {
		sendError(w, http.StatusBadRequest, "avatar too large (max 5MB)")
		return
	}

	// Validate image format
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		sendError(w, http.StatusBadRequest, "only image files are allowed")
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := "avatar_" + strconv.FormatUint(userID, 10) + "_" + uuid.New().String() + ext
	filePath := filepath.Join(h.cfg.File.Path, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		logger.Errorf("Failed to create avatar: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to save avatar")
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		logger.Errorf("Failed to copy avatar: %v", err)
		sendError(w, http.StatusInternalServerError, "failed to save avatar")
		return
	}

	// Generate avatar URL
	avatarURL := "/uploads/" + filename

	// Here you would update the user's avatar_url in the database
	// For now, we just return the URL

	logger.Infof("User %d uploaded avatar: %s", userID, filename)

	response := map[string]interface{}{
		"avatar_url": avatarURL,
	}

	sendSuccess(w, http.StatusOK, response)
}

func getFileType(contentType string) string {
	if strings.HasPrefix(contentType, "image/") {
		return "image"
	}
	if strings.HasPrefix(contentType, "video/") {
		return "video"
	}
	if strings.HasPrefix(contentType, "audio/") {
		return "audio"
	}
	if strings.HasPrefix(contentType, "application/pdf") {
		return "pdf"
	}
	return "document"
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
