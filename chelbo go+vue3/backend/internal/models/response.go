package models

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
	}
}

func NewMessageResponse(message string) *APIResponse {
	return &APIResponse{
		Success: true,
		Message: message,
	}
}

func NewErrorResponse(err string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error:   err,
	}
}
