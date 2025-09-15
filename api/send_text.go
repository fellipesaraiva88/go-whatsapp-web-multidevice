package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// SendText sends a text message
func SendText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Validate required fields
	if req.Phone == "" || req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Phone and message are required"})
		return
	}

	// Simulate message sending
	messageID := generateMessageID()
	
	// Store message in database
	service := NewWhatsAppService()
	if service != nil {
		storeMessage(service, req.Phone, "text", req.Message, messageID)
	}

	response := MessageResponse{
		Success:   true,
		MessageID: messageID,
		Message:   "Text message sent successfully",
		Timestamp: time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}