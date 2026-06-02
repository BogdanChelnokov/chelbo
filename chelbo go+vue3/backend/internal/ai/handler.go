package ai

import (
	"encoding/json"
	"net/http"

	"chelbo/backend/internal/auth"
	"chelbo/backend/internal/models"
	"chelbo/backend/internal/pkg/logger"
)

type Handler struct {
	assistant *Assistant
}

type AskRequest struct {
	Question string `json:"question"`
}

type TranslateRequest struct {
	Text       string `json:"text"`
	TargetLang string `json:"target_lang"`
}

type SummarizeRequest struct {
	Messages []string `json:"messages"`
}

func NewHandler(enabled, mockResponses bool) *Handler {
	return &Handler{
		assistant: NewAssistant(enabled, mockResponses),
	}
}

// Ask handles AI questions
func (h *Handler) Ask(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req AskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Question == "" {
		sendError(w, http.StatusBadRequest, "question is required")
		return
	}

	logger.Infof("User %d asked AI: %s", userID, req.Question)

	response := h.assistant.GenerateResponse(req.Question)

	sendSuccess(w, http.StatusOK, map[string]string{
		"answer": response,
	})
}

// Translate translates a message
func (h *Handler) Translate(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req TranslateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Text == "" {
		sendError(w, http.StatusBadRequest, "text is required")
		return
	}

	if req.TargetLang == "" {
		req.TargetLang = "ru"
	}

	logger.Infof("User %d requested translation to %s", userID, req.TargetLang)

	translated := h.assistant.TranslateMessage(req.Text, req.TargetLang)

	sendSuccess(w, http.StatusOK, map[string]string{
		"original":    req.Text,
		"translated":  translated,
		"target_lang": req.TargetLang,
	})
}

// Summarize generates a summary of chat messages
func (h *Handler) Summarize(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		sendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req SummarizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Messages) == 0 {
		sendError(w, http.StatusBadRequest, "messages are required")
		return
	}

	logger.Infof("User %d requested summary of %d messages", userID, len(req.Messages))

	summary := h.assistant.GetChatSummary(req.Messages)

	sendSuccess(w, http.StatusOK, map[string]string{
		"summary":       summary,
		"message_count": string(rune(len(req.Messages))),
	})
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
