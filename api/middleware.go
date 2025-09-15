package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Rate limiting configuration
type RateLimiter struct {
	requests map[string][]int64
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

var (
	globalRateLimiter = NewRateLimiter(100, time.Hour)     // 100 requests per hour
	authRateLimiter   = NewRateLimiter(5, time.Minute*15)  // 5 auth attempts per 15 minutes
)

// Security headers configuration
var securityHeaders = map[string]string{
	"X-Content-Type-Options":    "nosniff",
	"X-Frame-Options":          "DENY",
	"X-XSS-Protection":         "1; mode=block",
	"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	"Referrer-Policy":          "strict-origin-when-cross-origin",
	"Content-Security-Policy":  "default-src 'self'",
}

// Context keys for request data
type contextKey string

const (
	UserContextKey contextKey = "user"
	ClaimsContextKey contextKey = "claims"
)

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]int64),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if the request should be allowed
func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now().Unix()
	windowStart := now - int64(rl.window.Seconds())

	// Get or create request history for this identifier
	if _, exists := rl.requests[identifier]; !exists {
		rl.requests[identifier] = make([]int64, 0)
	}

	// Remove old requests outside the window
	requests := rl.requests[identifier]
	validRequests := make([]int64, 0)
	for _, reqTime := range requests {
		if reqTime > windowStart {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if limit exceeded
	if len(validRequests) >= rl.limit {
		rl.requests[identifier] = validRequests
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[identifier] = validRequests

	return true
}

// SecurityMiddleware adds security headers to all responses
func SecurityMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		for header, value := range securityHeaders {
			w.Header().Set(header, value)
		}

		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// RateLimitMiddleware applies rate limiting
func RateLimitMiddleware(limiter *RateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get client identifier (IP address)
			clientIP := getClientIP(r)
			
			if !limiter.Allow(clientIP) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "Rate limit exceeded",
					"message": "Too many requests. Please try again later.",
					"retry_after": int(limiter.window.Seconds()),
				})
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Extract token from header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Authorization header required",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid token format. Use 'Bearer <token>'",
			})
			return
		}

		// Validate token
		claims, err := validateJWTToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid or expired token",
			})
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		ctx = context.WithValue(ctx, UserContextKey, &User{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// AdminMiddleware requires admin role
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(ClaimsContextKey).(*Claims)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Authentication required",
			})
			return
		}

		if claims.Role != "admin" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Admin access required",
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}

// LoggingMiddleware logs requests
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a custom ResponseWriter to capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}
		
		next.ServeHTTP(lrw, r)
		
		duration := time.Since(start)
		
		// Log request details (in production, use proper logging library)
		fmt.Printf("[%s] %s %s %d %v %s\n",
			time.Now().Format(time.RFC3339),
			r.Method,
			r.RequestURI,
			lrw.statusCode,
			duration,
			getClientIP(r),
		)
	}
}

// Custom ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Helper functions
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (proxy/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check for X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Use RemoteAddr as fallback
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// GetUserFromContext extracts user from request context
func GetUserFromContext(r *http.Request) *User {
	if user, ok := r.Context().Value(UserContextKey).(*User); ok {
		return user
	}
	return nil
}

// GetClaimsFromContext extracts claims from request context
func GetClaimsFromContext(r *http.Request) *Claims {
	if claims, ok := r.Context().Value(ClaimsContextKey).(*Claims); ok {
		return claims
	}
	return nil
}

// ChainMiddleware combines multiple middleware functions
func ChainMiddleware(middlewares ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(final http.HandlerFunc) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}