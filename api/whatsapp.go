package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/supabase-community/postgrest-go"
)

type WhatsAppService struct {
	supabase *postgrest.Client
}

type QRCodeResponse struct {
	QRCode    string `json:"qr_code"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type ConnectionStatus struct {
	Connected   bool   `json:"connected"`
	Phone       string `json:"phone,omitempty"`
	DeviceID    string `json:"device_id,omitempty"`
	LastSeen    int64  `json:"last_seen,omitempty"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

type Session struct {
	ID          int    `json:"id"`
	JID         string `json:"jid"`
	DeviceID    int    `json:"device_id"`
	Platform    string `json:"platform"`
	BusinessName string `json:"business_name,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func NewWhatsAppService() *WhatsAppService {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	
	if supabaseURL == "" || supabaseKey == "" {
		return nil
	}
	
	client := postgrest.NewClient(supabaseURL+"/rest/v1", supabaseKey, nil)
	if client == nil {
		return nil
	}
	
	return &WhatsAppService{
		supabase: client,
	}
}

// QRCode generates a QR code for WhatsApp login
func QRCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// For now, simulate QR code generation
	// In a real implementation, this would use whatsmeow to generate actual QR codes
	response := QRCodeResponse{
		QRCode:    "2@BQwbZF9jNzY1NDMyMTEwMjMsNTU1LDEsWUhyR1hUOHRrYnd6TG5uVnlqZGlBRWFkUUZMdDBaZlE9",
		Status:    "waiting",
		Message:   "Scan this QR code with WhatsApp to connect",
		Timestamp: time.Now().Unix(),
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Status checks the connection status
func Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	service := NewWhatsAppService()
	if service == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to initialize WhatsApp service"})
		return
	}

	// Check for active sessions in database
	var sessions []Session
	result, err := service.supabase.From("whatsapp_sessions").Select("*").Execute()
	
	if err == nil {
		json.Unmarshal(result.Data, &sessions)
	}
	
	var status ConnectionStatus
	if err != nil || len(sessions) == 0 {
		status = ConnectionStatus{
			Connected: false,
			Status:    "disconnected",
			Message:   "No active WhatsApp session found. Please scan QR code to connect.",
		}
	} else {
		// Get the most recent session
		session := sessions[len(sessions)-1]
		status = ConnectionStatus{
			Connected: true,
			Phone:     session.JID,
			DeviceID:  fmt.Sprintf("%d", session.DeviceID),
			Status:    "connected",
			Message:   "WhatsApp is connected and ready",
			LastSeen:  time.Now().Unix(),
		}
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// Logout disconnects from WhatsApp
func Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	service := NewWhatsAppService()
	if service == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to initialize WhatsApp service"})
		return
	}

	// Delete all sessions from database
	_, err := service.supabase.From("whatsapp_sessions").Delete().Execute()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to logout"})
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Successfully logged out from WhatsApp",
		"timestamp": time.Now().Unix(),
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SaveSession saves a WhatsApp session to database
func (s *WhatsAppService) SaveSession(ctx context.Context, jid string, deviceID int, platform string) error {
	session := map[string]interface{}{
		"jid":       jid,
		"device_id": deviceID,
		"platform":  platform,
		"created_at": time.Now().UTC(),
		"updated_at": time.Now().UTC(),
	}
	
	_, err := s.supabase.From("whatsapp_sessions").Insert(session).Execute()
	return err
}

// Handler routes WhatsApp requests based on endpoint parameter
func Handler(w http.ResponseWriter, r *http.Request) {
	// Apply global middleware first
	GlobalMiddleware(func(w http.ResponseWriter, r *http.Request) {
		endpoint := r.URL.Query().Get("endpoint")
		
		switch endpoint {
		case "qr":
			QRCode(w, r)
		case "status":
			Status(w, r)
		case "logout":
			Logout(w, r)
		default:
			http.Error(w, "Invalid endpoint", http.StatusNotFound)
		}
	})(w, r)
}
