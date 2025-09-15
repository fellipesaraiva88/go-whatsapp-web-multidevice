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
	
	service.supabase.From("chat_storage").Insert(message).Execute()
}