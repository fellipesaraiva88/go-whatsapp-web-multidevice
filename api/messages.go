package handler

import (
	"net/http"
)

func MessagesHandler(w http.ResponseWriter, r *http.Request) {
	msgType := r.URL.Query().Get("type")

	switch msgType {
	case "text":
		SendText(w, r)
	case "image":
		SendImage(w, r)
	case "audio":
		SendAudio(w, r)
	case "file":
		SendFile(w, r)
	case "contact":
		SendContact(w, r)
	case "location":
		SendLocation(w, r)
	case "poll":
		SendPoll(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Message type not found"))
	}
}

package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type SendMessageRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Message  string `json:"message" validate:"required"`
	Duration int    `json:"duration,omitempty"` // For disappearing messages
}

type SendImageRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Image    string `json:"image,omitempty"`     // Base64 or URL
	ImageURL string `json:"image_url,omitempty"` // Alternative URL
	Caption  string `json:"caption,omitempty"`
}

type SendAudioRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Audio    string `json:"audio,omitempty"`     // Base64 or URL
	AudioURL string `json:"audio_url,omitempty"` // Alternative URL
	Duration int    `json:"duration,omitempty"`
}

type SendFileRequest struct {
	Phone    string `json:"phone" validate:"required"`
	File     string `json:"file,omitempty"`     // Base64 or URL
	FileURL  string `json:"file_url,omitempty"` // Alternative URL
	Filename string `json:"filename,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

type SendContactRequest struct {
	Phone        string `json:"phone" validate:"required"`
	ContactName  string `json:"contact_name" validate:"required"`
	ContactPhone string `json:"contact_phone" validate:"required"`
}

type SendLocationRequest struct {
	Phone     string  `json:"phone" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

type SendPollRequest struct {
	Phone      string   `json:"phone" validate:"required"`
	Question   string   `json:"question" validate:"required"`
	Options    []string `json:"options" validate:"required"`
	MaxAnswers int      `json:"max_answers,omitempty"`
}

type MessageResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

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

// SendImage sends an image message
func SendImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req SendImageRequest
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

	if req.Image == "" && req.ImageURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Image or image_url is required"})
		return
	}

	// Simulate image sending
	messageID := generateMessageID()
	
	// Store message in database
	service := NewWhatsAppService()
	if service != nil {
		storeMessage(service, req.Phone, "image", req.Caption, messageID)
	}

	response := MessageResponse{
		Success:   true,
		MessageID: messageID,
		Message:   "Image sent successfully",
		Timestamp: time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

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

// SendFile sends a document/file
func SendFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req SendFileRequest
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

	if req.File == "" && req.FileURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "File or file_url is required"})
		return
	}

	// Simulate file sending
	messageID := generateMessageID()
	
	// Store message in database
	service := NewWhatsAppService()
	if service != nil {
		storeMessage(service, req.Phone, "file", req.Caption, messageID)
	}

	response := MessageResponse{
		Success:   true,
		MessageID: messageID,
		Message:   "File sent successfully",
		Timestamp: time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

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

// SendPoll sends a poll/vote
func SendPoll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var req SendPollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Validate required fields
	if req.Phone == "" || req.Question == "" || len(req.Options) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Phone, question, and options are required"})
		return
	}

	// Simulate poll sending
	messageID := generateMessageID()
	
	// Store message in database
	service := NewWhatsAppService()
	if service != nil {
		storeMessage(service, req.Phone, "poll", req.Question, messageID)
	}

	response := MessageResponse{
		Success:   true,
		MessageID: messageID,
		Message:   "Poll sent successfully",
		Timestamp: time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper functions
func generateMessageID() string {
	return "msg_" + strings.ReplaceAll(time.Now().Format("20060102150405.000"), ".", "")
}

func storeMessage(service *WhatsAppService, phone, msgType, content, messageID string) {
	if service == nil {
		return
	}
	
	message := map[string]interface{}{
		"jid":         phone,
		"message_id":  messageID,
		"message_data": map[string]interface{}{
			"type":    msgType,
			"content": content,
			"sent":    true,
		},
		"timestamp": time.Now().UTC(),
	}
	
	service.supabase.DB.From("chat_storage").Insert(message).Execute()
}