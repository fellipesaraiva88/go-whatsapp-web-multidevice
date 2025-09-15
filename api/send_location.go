package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// SendLocation sends a location
func SendLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req SendLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Validate required fields
	if req.Phone == "" || req.Latitude == 0 || req.Longitude == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Phone, latitude, and longitude are required"})
		return
	}

	// Simulate location sending
	messageID := generateMessageID()
	
	// Store message in database
	service := NewWhatsAppService()
	if service != nil {
		storeMessage(service, req.Phone, "location", req.Name, messageID)
	}

	response := MessageResponse{
		Success:   true,
		MessageID: messageID,
		Message:   "Location sent successfully",
		Timestamp: time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}