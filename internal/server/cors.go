package server

import (
	"net/http"
	"strings"
)

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	AllowedOrigins []string
}

// NewCORSConfig creates a CORSConfig from a comma-separated origins string.
// If the string is empty or "*", defaults to allow all origins.
func NewCORSConfig(origins string) *CORSConfig {
	if origins == "" || origins == "*" {
		return &CORSConfig{AllowedOrigins: []string{"*"}}
	}
	parts := strings.Split(origins, ",")
	trimmed := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			trimmed = append(trimmed, p)
		}
	}
	if len(trimmed) == 0 {
		return &CORSConfig{AllowedOrigins: []string{"*"}}
	}
	return &CORSConfig{AllowedOrigins: trimmed}
}

// corsMiddleware returns an http.Handler that adds CORS headers to responses
// and handles OPTIONS preflight requests.
func corsMiddleware(cfg *CORSConfig, next http.Handler) http.Handler {
	allowOrigin := strings.Join(cfg.AllowedOrigins, ", ")
	if allowOrigin == "" {
		allowOrigin = "*"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Mcp-Session-Id")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
