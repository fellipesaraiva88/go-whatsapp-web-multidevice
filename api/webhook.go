package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type WebhookEvent struct {
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
	From      string                 `json:"from"`
	To        string                 `json:"to,omitempty"`
	MessageID string                 `json:"message_id"`
	Data      map[string]interface{} `json:"data"`
	User      string                 `json:"user,omitempty"`
}

type IncomingMessage struct {
	From        string                 `json:"from"`
	MessageID   string                 `json:"message_id"`
	MessageType string                 `json:"message_type"`
	Content     string                 `json:"content,omitempty"`
	Caption     string                 `json:"caption,omitempty"`
	MediaURL    string                 `json:"media_url,omitempty"`
	Timestamp   int64                  `json:"timestamp"`
	IsGroup     bool                   `json:"is_group"`
	GroupName   string                 `json:"group_name,omitempty"`
	SenderName  string                 `json:"sender_name,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

type WebhookResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	EventID     string `json:"event_id,omitempty"`
	ProcessedAt int64  `json:"processed_at"`
}

// ReceiveWebhook processes incoming webhooks (simulated for now)
func ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(WebhookResponse{
			Success: false,
			Message: "Method not allowed",
			ProcessedAt: time.Now().Unix(),
		})
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(WebhookResponse{
			Success: false,
			Message: "Failed to read request body",
			ProcessedAt: time.Now().Unix(),
		})
		return
	}

	// Verify webhook signature if secret is configured
	if !verifyWebhookSignature(r.Header.Get("X-Hub-Signature-256"), body) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(WebhookResponse{
			Success: false,
			Message: "Invalid webhook signature",
			ProcessedAt: time.Now().Unix(),
		})
		return
	}

	// Parse webhook payload
	var message IncomingMessage
	if err := json.Unmarshal(body, &message); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(WebhookResponse{
			Success: false,
			Message: "Invalid JSON payload",
			ProcessedAt: time.Now().Unix(),
		})
		return
	}

	// Process the incoming message
	eventID, err := processIncomingMessage(message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(WebhookResponse{
			Success: false,
			Message: "Failed to process message: " + err.Error(),
			ProcessedAt: time.Now().Unix(),
		})
		return
	}

	// Send success response
	response := WebhookResponse{
		Success:     true,
		Message:     "Message processed successfully",
		EventID:     eventID,
		ProcessedAt: time.Now().Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SendWebhook sends webhook events to configured URLs
func SendWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Get user from context (requires authentication)
	user := GetUserFromContext(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	var event WebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	// Add user information to event
	event.User = user.Username
	event.Timestamp = time.Now().Unix()

	// Send to configured webhook URLs
	results, err := sendToWebhookURLs(event)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to send webhooks"})
		return
	}

	response := map[string]interface{}{
		"success":       true,
		"event_id":      generateMessageID(),
		"webhook_urls":  len(results),
		"sent_at":       time.Now().Unix(),
		"results":       results,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ManageWebhooks allows managing webhook URLs (admin only)
func ManageWebhooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	user := GetUserFromContext(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Get current webhook configuration
		webhooks := getWebhookURLs()
		response := map[string]interface{}{
			"webhooks": webhooks,
			"count":    len(webhooks),
			"secret_configured": os.Getenv("WHATSAPP_WEBHOOK_SECRET") != "",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

	case http.MethodPost:
		// Add webhook URL
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
			return
		}

		if req.URL == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "URL is required"})
			return
		}

		// In a real implementation, this would update the webhook configuration
		response := map[string]interface{}{
			"success": true,
			"message": "Webhook URL added successfully",
			"url":     req.URL,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

// Helper functions

func verifyWebhookSignature(signature string, payload []byte) bool {
	secret := os.Getenv("WHATSAPP_WEBHOOK_SECRET")
	if secret == "" {
		secret = "super-secret-webhook-key"
	}

	if signature == "" {
		return false
	}

	// Remove 'sha256=' prefix if present
	signature = strings.TrimPrefix(signature, "sha256=")

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func processIncomingMessage(message IncomingMessage) (string, error) {
	service := NewWhatsAppService()
	if service == nil {
		return "", fmt.Errorf("service unavailable")
	}

	// Generate event ID
	eventID := generateMessageID()

	// Store incoming message in database
	incomingData := map[string]interface{}{
		"jid":        message.From,
		"message_id": message.MessageID,
		"message_data": map[string]interface{}{
			"type":         message.MessageType,
			"content":      message.Content,
			"caption":      message.Caption,
			"media_url":    message.MediaURL,
			"is_group":     message.IsGroup,
			"group_name":   message.GroupName,
			"sender_name":  message.SenderName,
			"received":     true,
			"event_id":     eventID,
		},
		"timestamp": time.Now().UTC(),
	}

	_, err := service.supabase.DB.From("chat_storage").Insert(incomingData).Execute()
	if err != nil {
		return eventID, fmt.Errorf("failed to store message: %v", err)
	}

	// Create webhook event
	webhookEvent := WebhookEvent{
		Type:      "message_received",
		Timestamp: time.Now().Unix(),
		From:      message.From,
		MessageID: message.MessageID,
		Data: map[string]interface{}{
			"message_type": message.MessageType,
			"content":      message.Content,
			"is_group":     message.IsGroup,
			"event_id":     eventID,
		},
	}

	// Send to webhook URLs (fire and forget)
	go func() {
		_, err := sendToWebhookURLs(webhookEvent)
		if err != nil {
			fmt.Printf("Failed to send webhook: %v\n", err)
		}
	}()

	return eventID, nil
}

func sendToWebhookURLs(event WebhookEvent) ([]map[string]interface{}, error) {
	urls := getWebhookURLs()
	if len(urls) == 0 {
		return nil, fmt.Errorf("no webhook URLs configured")
	}

	var results []map[string]interface{}
	
	for _, url := range urls {
		result := sendSingleWebhook(url, event)
		results = append(results, result)
	}

	return results, nil
}

func sendSingleWebhook(url string, event WebhookEvent) map[string]interface{} {
	result := map[string]interface{}{
		"url":       url,
		"success":   false,
		"timestamp": time.Now().Unix(),
	}

	// Prepare payload
	payload, err := json.Marshal(event)
	if err != nil {
		result["error"] = "Failed to marshal event"
		return result
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		result["error"] = "Failed to create request"
		return result
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "WhatsApp-API-Webhook/1.0")

	// Add signature header
	secret := os.Getenv("WHATSAPP_WEBHOOK_SECRET")
	if secret == "" {
		secret = "super-secret-webhook-key"
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))
	req.Header.Set("X-Hub-Signature-256", "sha256="+signature)

	// Send request with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		result["error"] = "Request failed: " + err.Error()
		return result
	}
	defer resp.Body.Close()

	result["status_code"] = resp.StatusCode
	result["success"] = resp.StatusCode >= 200 && resp.StatusCode < 300

	if !result["success"].(bool) {
		result["error"] = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return result
}

func getWebhookURLs() []string {
	webhookURLs := os.Getenv("WHATSAPP_WEBHOOK")
	if webhookURLs == "" {
		return []string{}
	}

	// Split by comma and trim whitespace
	urls := strings.Split(webhookURLs, ",")
	var cleanURLs []string
	for _, url := range urls {
		if trimmed := strings.TrimSpace(url); trimmed != "" {
			cleanURLs = append(cleanURLs, trimmed)
		}
	}

	return cleanURLs
}

// Handler routes webhook requests based on endpoint parameter
func Handler(w http.ResponseWriter, r *http.Request) {
	// Apply global middleware first
	GlobalMiddleware(func(w http.ResponseWriter, r *http.Request) {
		endpoint := r.URL.Query().Get("endpoint")
		
		switch endpoint {
		case "receive":
			ReceiveWebhook(w, r)
		case "send":
			// Send webhook requires authentication
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				SendWebhook(w, r)
			})(w, r)
		case "manage":
			// Manage webhooks requires authentication
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				ManageWebhooks(w, r)
			})(w, r)
		default:
			http.Error(w, "Invalid endpoint", http.StatusNotFound)
		}
	})(w, r)
}
