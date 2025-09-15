package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

// Profile returns user profile information (protected endpoint)
func Profile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Get user from context (added by AuthMiddleware)
	user := GetUserFromContext(r)
	claims := GetClaimsFromContext(r)

	if user == nil || claims == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	// Get WhatsApp connection status
	service := NewWhatsAppService()
	var sessions []Session
	var whatsappStatus string = "disconnected"
	
	if service != nil {
	result, err := service.supabase.From("whatsapp_sessions").Select("*").Execute()
	if err == nil {
		json.Unmarshal(result.Data, &sessions)
	}
		if err == nil && len(sessions) > 0 {
			whatsappStatus = "connected"
		}
	}

	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
			"created":  user.Created,
		},
		"token_info": map[string]interface{}{
			"issued_at": claims.IssuedAt.Time.Unix(),
			"expires":   claims.ExpiresAt.Time.Unix(),
			"issuer":    claims.Issuer,
		},
		"whatsapp": map[string]interface{}{
			"status":    whatsappStatus,
			"sessions":  len(sessions),
		},
		"permissions": getUserPermissions(user.Role),
		"timestamp":   time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// AdminDashboard returns admin-only information (admin-protected endpoint)
func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	user := GetUserFromContext(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	// Get system statistics
	service := NewWhatsAppService()
	var stats map[string]interface{} = map[string]interface{}{
		"total_sessions":  0,
		"total_messages":  0,
		"active_webhooks": 0,
		"uptime":         "0h 0m",
		"last_activity":   time.Now().Unix(),
	}

	if service != nil {
		var sessions []Session
		var messages []map[string]interface{}
		
	result1, _ := service.supabase.From("whatsapp_sessions").Select("*").Execute()
	result2, _ := service.supabase.From("chat_storage").Select("*").Execute()
	
	if result1 != nil {
		json.Unmarshal(result1.Data, &sessions)
	}
	if result2 != nil {
		json.Unmarshal(result2.Data, &messages)
	}
		
		stats["total_sessions"] = len(sessions)
		stats["total_messages"] = len(messages)
	}

	response := map[string]interface{}{
		"admin":      user.Username,
		"statistics": stats,
		"system_info": map[string]interface{}{
			"version":    "1.0.0",
			"build_date": "2025-09-15",
			"environment": "production",
			"features": []string{
				"jwt_auth",
				"rate_limiting", 
				"webhooks",
				"message_storage",
				"admin_panel",
			},
		},
		"rate_limits": map[string]interface{}{
			"global": map[string]interface{}{
				"limit":  100,
				"window": "1h",
			},
			"auth": map[string]interface{}{
				"limit":  5,
				"window": "15m",
			},
		},
		"timestamp": time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ProtectedSendMessage is a protected version of send message with user logging
func ProtectedSendMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	user := GetUserFromContext(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
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

	// Generate message ID and send
	messageID := generateMessageID()
	
	// Store message with user information
	service := NewWhatsAppService()
	if service != nil {
		message := map[string]interface{}{
			"jid":         req.Phone,
			"message_id":  messageID,
			"message_data": map[string]interface{}{
				"type":     "text",
				"content":  req.Message,
				"sent":     true,
				"sent_by":  user.Username,
				"user_id":  user.ID,
			},
			"timestamp": time.Now().UTC(),
		}
	service.supabase.From("chat_storage").Insert(message).Execute()
	}

	response := MessageResponse{
		Success:   true,
		MessageID: messageID,
		Message:   "Text message sent successfully by " + user.Username,
		Timestamp: time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetMessageHistory returns message history for authenticated users
func GetMessageHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	user := GetUserFromContext(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	service := NewWhatsAppService()
	if service == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Service unavailable"})
		return
	}

	// Get query parameters
	phone := r.URL.Query().Get("phone")
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "50" // Default limit
	}

	// Build query
	query := service.supabase.From("chat_storage").Select("*").Order("timestamp", false).Limit(50)
	
	if phone != "" {
		query = query.Eq("jid", phone)
	}

	var messages []map[string]interface{}
	result, err := query.Execute()
	if err == nil && result != nil {
		json.Unmarshal(result.Data, &messages)
	}
	
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch messages"})
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"messages":  messages,
		"count":     len(messages),
		"requested_by": user.Username,
		"timestamp": time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Handler routes protected requests based on endpoint parameter
func Handler(w http.ResponseWriter, r *http.Request) {
	// Apply global middleware and auth middleware
	GlobalMiddleware(func(w http.ResponseWriter, r *http.Request) {
		AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
			endpoint := r.URL.Query().Get("endpoint")
			
			switch endpoint {
			case "profile":
				Profile(w, r)
			case "dashboard":
				AdminMiddleware(func(w http.ResponseWriter, r *http.Request) {
					AdminDashboard(w, r)
				})(w, r)
			case "send":
				ProtectedSendMessage(w, r)
			case "history":
				GetMessageHistory(w, r)
			default:
				http.Error(w, "Invalid endpoint", http.StatusNotFound)
			}
		})(w, r)
	})(w, r)
}

// Helper function to get user permissions
func getUserPermissions(role string) []string {
	switch role {
	case "admin":
		return []string{
			"send_message",
			"view_messages", 
			"manage_sessions",
			"admin_dashboard",
			"manage_users",
			"view_logs",
		}
	case "user":
		return []string{
			"send_message",
			"view_messages",
		}
	default:
		return []string{}
	}
}