package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// SendContact sends a contact card
func SendContact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req SendContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Validate required fields
	if req.Phone == "" || req.ContactName == "" || req.ContactPhone == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Phone, contact_name, and contact_phone are required"})
		return
	}

	// Simulate contact sending
	messageID := generateMessageID()
	
	// Store message in database
	service := NewWhatsAppService()
	if service != nil {
		storeMessage(service, req.Phone, "contact", req.ContactName, messageID)
	}

	response := MessageResponse{
		Success:   true,
		MessageID: messageID,
		Message:   "Contact sent successfully",
		Timestamp: time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}