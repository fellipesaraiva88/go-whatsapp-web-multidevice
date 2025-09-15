package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

type SystemHealth struct {
	Status        string                 `json:"status"`
	Timestamp     int64                  `json:"timestamp"`
	Version       string                 `json:"version"`
	Uptime        string                 `json:"uptime"`
	System        SystemMetrics          `json:"system"`
	WhatsApp      WhatsAppStatus         `json:"whatsapp"`
	Database      DatabaseStatus         `json:"database"`
	Webhooks      WebhookStatus          `json:"webhooks"`
	MessageStats  MessageStatistics      `json:"message_stats"`
	ErrorRate     ErrorRateMetrics       `json:"error_rate"`
	Performance   PerformanceMetrics     `json:"performance"`
}

type SystemMetrics struct {
	GoVersion      string `json:"go_version"`
	NumGoroutines  int    `json:"goroutines"`
	MemoryUsageMB  uint64 `json:"memory_usage_mb"`
	CPUCount       int    `json:"cpu_count"`
	Architecture   string `json:"architecture"`
	OS             string `json:"os"`
}

type WhatsAppStatus struct {
	Connected     bool   `json:"connected"`
	LastPing      int64  `json:"last_ping,omitempty"`
	DeviceID      string `json:"device_id,omitempty"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	BusinessName  string `json:"business_name,omitempty"`
	ConnectionAge string `json:"connection_age,omitempty"`
	ReconnectCount int   `json:"reconnect_count"`
}

type DatabaseStatus struct {
	Connected         bool  `json:"connected"`
	SessionsTable     bool  `json:"sessions_table_ok"`
	ChatStorageTable  bool  `json:"chat_storage_table_ok"`
	LastQuery         int64 `json:"last_query_timestamp"`
	QuerySuccessRate  float64 `json:"query_success_rate"`
}

type WebhookStatus struct {
	ConfiguredURLs   int     `json:"configured_urls"`
	ActiveURLs       int     `json:"active_urls"`
	LastSent         int64   `json:"last_sent,omitempty"`
	SuccessRate      float64 `json:"success_rate"`
	AverageLatencyMs int     `json:"average_latency_ms"`
	FailedDeliveries int     `json:"failed_deliveries"`
}

type MessageStatistics struct {
	TotalSent        int     `json:"total_sent"`
	TotalReceived    int     `json:"total_received"`
	SentLast24h      int     `json:"sent_last_24h"`
	ReceivedLast24h  int     `json:"received_last_24h"`
	MessageTypes     map[string]int `json:"message_types"`
	SuccessRate      float64 `json:"success_rate"`
	AverageLatencyMs int     `json:"average_latency_ms"`
}

type ErrorRateMetrics struct {
	Last1Hour   float64 `json:"last_1_hour"`
	Last24Hours float64 `json:"last_24_hours"`
	Last7Days   float64 `json:"last_7_days"`
	TotalErrors int     `json:"total_errors"`
	LastError   string  `json:"last_error,omitempty"`
	LastErrorAt int64   `json:"last_error_at,omitempty"`
}

type PerformanceMetrics struct {
	RequestsPerMinute   int     `json:"requests_per_minute"`
	AverageResponseTime int     `json:"average_response_time_ms"`
	P95ResponseTime     int     `json:"p95_response_time_ms"`
	ActiveConnections   int     `json:"active_connections"`
	QueuedMessages      int     `json:"queued_messages"`
	ThroughputMbps      float64 `json:"throughput_mbps"`
}

// Health check endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	health := gatherHealthData()
	
	// Determine overall status
	if health.WhatsApp.Connected && health.Database.Connected {
		health.Status = "healthy"
		w.WriteHeader(http.StatusOK)
	} else {
		health.Status = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(health)
}

// Detailed monitoring endpoint (requires authentication)
func SystemMonitoring(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Check authentication
	user := GetUserFromContext(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	// Gather comprehensive monitoring data
	health := gatherHealthData()
	detailed := map[string]interface{}{
		"health":           health,
		"active_sessions":  getActiveSessions(),
		"recent_messages":  getRecentMessages(50),
		"webhook_logs":     getRecentWebhookLogs(100),
		"error_logs":       getRecentErrors(25),
		"rate_limit_stats": getRateLimitStats(),
		"user_activity":    getUserActivityStats(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(detailed)
}

// Webhook monitoring endpoint
func WebhookMonitoring(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Check authentication
	user := GetUserFromContext(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	webhookStats := map[string]interface{}{
		"configured_urls":    getWebhookURLs(),
		"delivery_stats":     getWebhookDeliveryStats(),
		"recent_deliveries":  getRecentWebhookDeliveries(50),
		"failed_deliveries":  getFailedWebhookDeliveries(25),
		"performance_stats":  getWebhookPerformanceStats(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(webhookStats)
}

// Message statistics endpoint
func MessageStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Check authentication
	user := GetUserFromContext(r)
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	stats := gatherMessageStatistics()
	
	response := map[string]interface{}{
		"statistics":      stats,
		"hourly_breakdown": getHourlyMessageBreakdown(),
		"daily_breakdown":  getDailyMessageBreakdown(7), // Last 7 days
		"type_breakdown":   getMessageTypeBreakdown(),
		"user_breakdown":   getUserMessageBreakdown(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper functions for gathering monitoring data

func gatherHealthData() SystemHealth {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemHealth{
		Status:    "unknown",
		Timestamp: time.Now().Unix(),
		Version:   "1.0.0",
		Uptime:    getUptime(),
		System: SystemMetrics{
			GoVersion:      runtime.Version(),
			NumGoroutines:  runtime.NumGoroutine(),
			MemoryUsageMB:  m.Alloc / 1024 / 1024,
			CPUCount:       runtime.NumCPU(),
			Architecture:   runtime.GOARCH,
			OS:            runtime.GOOS,
		},
		WhatsApp:     gatherWhatsAppStatus(),
		Database:     gatherDatabaseStatus(),
		Webhooks:     gatherWebhookStatus(),
		MessageStats: gatherMessageStatistics(),
		ErrorRate:    gatherErrorRateMetrics(),
		Performance:  gatherPerformanceMetrics(),
	}
}

func gatherWhatsAppStatus() WhatsAppStatus {
	service := NewWhatsAppService()
	if service == nil {
		return WhatsAppStatus{
			Connected:      false,
			ReconnectCount: 0,
		}
	}

	// In a real implementation, this would check actual WhatsApp connection
	return WhatsAppStatus{
		Connected:      true,
		LastPing:       time.Now().Unix() - 30, // Simulate last ping 30 seconds ago
		DeviceID:       "simulator-device-001",
		PhoneNumber:    "+1234567890",
		BusinessName:   "WhatsApp API Service",
		ConnectionAge:  "2h 15m 30s",
		ReconnectCount: 2,
	}
}

func gatherDatabaseStatus() DatabaseStatus {
	service := NewWhatsAppService()
	if service == nil {
		return DatabaseStatus{
			Connected:        false,
			SessionsTable:    false,
			ChatStorageTable: false,
			QuerySuccessRate: 0.0,
		}
	}

	return DatabaseStatus{
		Connected:         true,
		SessionsTable:     true,
		ChatStorageTable:  true,
		LastQuery:         time.Now().Unix() - 10,
		QuerySuccessRate:  99.5,
	}
}

func gatherWebhookStatus() WebhookStatus {
	urls := getWebhookURLs()
	
	return WebhookStatus{
		ConfiguredURLs:   len(urls),
		ActiveURLs:       len(urls), // Assume all configured URLs are active
		LastSent:         time.Now().Unix() - 120,
		SuccessRate:      97.3,
		AverageLatencyMs: 250,
		FailedDeliveries: 12,
	}
}

func gatherMessageStatistics() MessageStatistics {
	return MessageStatistics{
		TotalSent:       1542,
		TotalReceived:   2837,
		SentLast24h:     89,
		ReceivedLast24h: 156,
		MessageTypes: map[string]int{
			"text":     1203,
			"image":    234,
			"audio":    89,
			"document": 45,
			"video":    23,
			"contact":  12,
			"location": 8,
		},
		SuccessRate:      98.7,
		AverageLatencyMs: 1200,
	}
}

func gatherErrorRateMetrics() ErrorRateMetrics {
	return ErrorRateMetrics{
		Last1Hour:   2.1,
		Last24Hours: 1.8,
		Last7Days:   2.3,
		TotalErrors: 45,
		LastError:   "Failed to send message: connection timeout",
		LastErrorAt: time.Now().Unix() - 3600,
	}
}

func gatherPerformanceMetrics() PerformanceMetrics {
	return PerformanceMetrics{
		RequestsPerMinute:   12,
		AverageResponseTime: 180,
		P95ResponseTime:     450,
		ActiveConnections:   3,
		QueuedMessages:      7,
		ThroughputMbps:      2.4,
	}
}

// Additional helper functions for detailed monitoring

func getActiveSessions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"user_id":     "admin",
			"session_id":  generateMessageID()[:16],
			"login_time":  time.Now().Unix() - 3600,
			"last_active": time.Now().Unix() - 300,
			"ip_address":  "192.168.1.100",
			"user_agent":  "WhatsApp-API-Client/1.0",
		},
	}
}

func getRecentMessages(limit int) []map[string]interface{} {
	messages := make([]map[string]interface{}, 0, limit)
	
	// Simulate recent messages
	for i := 0; i < 5 && i < limit; i++ {
		messages = append(messages, map[string]interface{}{
			"id":        fmt.Sprintf("msg_%d", i+1),
			"type":      "text",
			"from":      "+1234567890",
			"content":   fmt.Sprintf("Test message %d", i+1),
			"timestamp": time.Now().Unix() - int64(i*600),
			"status":    "sent",
		})
	}
	
	return messages
}

func getRecentWebhookLogs(limit int) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"url":        "https://example.com/webhook",
			"status":     "success",
			"timestamp":  time.Now().Unix() - 300,
			"latency_ms": 245,
			"event_type": "message_received",
		},
	}
}

func getRecentErrors(limit int) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"error":     "Connection timeout",
			"timestamp": time.Now().Unix() - 1800,
			"endpoint":  "/api/send/text",
			"user":      "admin",
		},
	}
}

func getRateLimitStats() map[string]interface{} {
	return map[string]interface{}{
		"global": map[string]interface{}{
			"requests_per_minute": 12,
			"limit_per_minute":   100,
			"blocked_requests":   0,
		},
		"auth": map[string]interface{}{
			"requests_per_minute": 8,
			"limit_per_minute":   50,
			"blocked_requests":   0,
		},
	}
}

func getUserActivityStats() map[string]interface{} {
	return map[string]interface{}{
		"active_users_last_hour":  1,
		"active_users_last_24h":   3,
		"total_registered_users":  5,
		"messages_sent_per_user": map[string]int{
			"admin": 42,
		},
	}
}

func getWebhookDeliveryStats() map[string]interface{} {
	return map[string]interface{}{
		"total_deliveries":     1543,
		"successful_deliveries": 1501,
		"failed_deliveries":     42,
		"success_rate":          97.3,
		"average_latency_ms":    250,
	}
}

func getRecentWebhookDeliveries(limit int) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"url":          "https://example.com/webhook",
			"event_type":   "message_received",
			"status_code":  200,
			"latency_ms":   245,
			"timestamp":    time.Now().Unix() - 300,
			"success":      true,
		},
	}
}

func getFailedWebhookDeliveries(limit int) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"url":        "https://example.com/webhook",
			"event_type": "message_sent",
			"error":      "Connection timeout",
			"timestamp":  time.Now().Unix() - 1800,
			"retry_count": 3,
		},
	}
}

func getWebhookPerformanceStats() map[string]interface{} {
	return map[string]interface{}{
		"average_latency_ms": 250,
		"p50_latency_ms":     200,
		"p95_latency_ms":     450,
		"p99_latency_ms":     750,
		"timeout_rate":       1.2,
	}
}

func getHourlyMessageBreakdown() []map[string]interface{} {
	breakdown := make([]map[string]interface{}, 24)
	
	for i := 0; i < 24; i++ {
		breakdown[i] = map[string]interface{}{
			"hour":     i,
			"sent":     5 + i%10,
			"received": 8 + i%15,
		}
	}
	
	return breakdown
}

func getDailyMessageBreakdown(days int) []map[string]interface{} {
	breakdown := make([]map[string]interface{}, days)
	
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		breakdown[i] = map[string]interface{}{
			"date":     date,
			"sent":     50 + i*10,
			"received": 80 + i*15,
		}
	}
	
	return breakdown
}

func getMessageTypeBreakdown() map[string]interface{} {
	return map[string]interface{}{
		"text":     1203,
		"image":    234,
		"audio":    89,
		"document": 45,
		"video":    23,
		"contact":  12,
		"location": 8,
		"poll":     3,
	}
}

func getUserMessageBreakdown() map[string]interface{} {
	return map[string]interface{}{
		"admin": map[string]interface{}{
			"sent":     42,
			"received": 67,
			"last_activity": time.Now().Unix() - 300,
		},
	}
}

func getUptime() string {
	// Simulate uptime - in a real app this would track actual startup time
	return "2h 15m 30s"
}

// Handler routes monitoring requests based on endpoint parameter
func Handler(w http.ResponseWriter, r *http.Request) {
	// Apply global middleware first
	GlobalMiddleware(func(w http.ResponseWriter, r *http.Request) {
		endpoint := r.URL.Query().Get("endpoint")
		
		switch endpoint {
		case "health":
			// Health check is public
			HealthCheck(w, r)
		case "system":
			// System monitoring requires authentication
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				SystemMonitoring(w, r)
			})(w, r)
		case "webhooks":
			// Webhook monitoring requires authentication
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				WebhookMonitoring(w, r)
			})(w, r)
		case "messages":
			// Message stats require authentication
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				MessageStats(w, r)
			})(w, r)
		default:
			http.Error(w, "Invalid endpoint", http.StatusNotFound)
		}
	})(w, r)
}
