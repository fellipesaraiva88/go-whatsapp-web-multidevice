package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// SendAudio sends an audio message
func SendAudio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req SendAudioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Validate required fields
	if req.Phone == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Phone is required"})
		return
	}

	if req.Audio == "" && req.AudioURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Audio or audio_url is required"})
		return
	}

	// Simulate audio sending
	messageID := generateMessageID()
	
	// Store message in database
	service := NewWhatsAppService()
	if service != nil {
		storeMessage(service, req.Phone, "audio", "Audio message", messageID)
	}

	response := MessageResponse{
		Success:   true,
		MessageID: messageID,
		Message:   "Audio sent successfully",
		Timestamp: time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}