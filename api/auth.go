package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT configuration
var (
	jwtSecret = []byte(getJWTSecret())
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type AuthRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Success      bool   `json:"success"`
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	Message      string `json:"message"`
	User         *User  `json:"user,omitempty"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Created  int64  `json:"created"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Login authenticates user and returns JWT token
func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "Invalid JSON",
		})
		return
	}

	// Validate credentials
	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "Username and password are required",
		})
		return
	}

	// Check credentials (in production, this should query a secure database)
	if !validateCredentials(req.Username, req.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// Generate tokens
	token, refreshToken, expiresIn, err := generateTokens(req.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "Failed to generate tokens",
		})
		return
	}

	// Create user object
	user := &User{
		ID:       generateUserID(req.Username),
		Username: req.Username,
		Role:     getUserRole(req.Username),
		Created:  time.Now().Unix(),
	}

	response := AuthResponse{
		Success:      true,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		Message:      "Authentication successful",
		User:         user,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RefreshToken refreshes an existing JWT token
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "Invalid JSON",
		})
		return
	}

	// Validate refresh token
	claims, err := validateRefreshToken(req.RefreshToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "Invalid refresh token",
		})
		return
	}

	// Generate new tokens
	token, refreshToken, expiresIn, err := generateTokens(claims.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "Failed to generate tokens",
		})
		return
	}

	response := AuthResponse{
		Success:      true,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		Message:      "Token refreshed successfully",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ValidateToken validates a JWT token and returns claims
func ValidateToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// Extract token from header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authorization header required"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token format"})
		return
	}

	// Validate token
	claims, err := validateJWTToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
		return
	}

	response := map[string]interface{}{
		"valid":     true,
		"user_id":   claims.UserID,
		"username":  claims.Username,
		"role":      claims.Role,
		"expires":   claims.ExpiresAt.Time.Unix(),
		"issued_at": claims.IssuedAt.Time.Unix(),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper functions
func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "whatsapp-api-secret-key-change-in-production"
	}
	return secret
}

func validateCredentials(username, password string) bool {
	// Get basic auth from environment or use defaults
	basicAuth := os.Getenv("APP_BASIC_AUTH")
	if basicAuth == "" {
		basicAuth = "admin:whatsapp2024"
	}

	// Parse multiple credentials (comma-separated)
	credentials := strings.Split(basicAuth, ",")
	for _, cred := range credentials {
		parts := strings.Split(strings.TrimSpace(cred), ":")
		if len(parts) == 2 && parts[0] == username && parts[1] == password {
			return true
		}
	}
	return false
}

func generateTokens(username string) (string, string, int64, error) {
	// Access token (1 hour)
	expiresIn := int64(3600) // 1 hour
	claims := &Claims{
		UserID:   generateUserID(username),
		Username: username,
		Role:     getUserRole(username),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiresIn) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "whatsapp-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", "", 0, err
	}

	// Refresh token (7 days)
	refreshClaims := &Claims{
		UserID:   generateUserID(username),
		Username: username,
		Role:     getUserRole(username),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "whatsapp-api",
		},
	}

	refreshTokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenJWT.SignedString(jwtSecret)
	if err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, expiresIn, nil
}

func validateJWTToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func validateRefreshToken(tokenString string) (*Claims, error) {
	// Same validation as JWT token since refresh token is also JWT
	return validateJWTToken(tokenString)
}

func generateUserID(username string) string {
	hash := sha256.Sum256([]byte(username + time.Now().String()))
	return base64.URLEncoding.EncodeToString(hash[:])[:16]
}

func getUserRole(username string) string {
	// Default role logic - can be expanded
	if username == "admin" {
		return "admin"
	}
	return "user"
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Handler routes auth requests based on endpoint parameter
func Handler(w http.ResponseWriter, r *http.Request) {
	// Apply global middleware first
	GlobalMiddleware(func(w http.ResponseWriter, r *http.Request) {
		endpoint := r.URL.Query().Get("endpoint")
		
		switch endpoint {
		case "login":
			Login(w, r)
		case "refresh":
			RefreshToken(w, r)
		case "validate":
			ValidateToken(w, r)
		default:
			http.Error(w, "Invalid endpoint", http.StatusNotFound)
		}
	})(w, r)
}
