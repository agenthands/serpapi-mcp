package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// contextKey is a custom type to avoid context key collisions.
type contextKey string

// apiKeyContextKey is the context key used to store the API key.
const apiKeyContextKey contextKey = "apiKey"

// APIKeyFromContext extracts the API key from the given context.
// Returns an empty string if no API key is present.
func APIKeyFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(apiKeyContextKey).(string); ok {
		return val
	}
	return ""
}

// ContextWithAPIKey returns a new context with the given API key stored.
// This is primarily useful for testing tool handlers that call APIKeyFromContext.
func ContextWithAPIKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, apiKeyContextKey, key)
}

// authErrorResponse is the JSON body returned for authentication failures.
type authErrorResponse struct {
	Error string `json:"error"`
}

// authOrPassthrough returns the auth middleware wrapping next, or a passthrough
// handler if disabled is true. Useful for testing without requiring API keys.
func authOrPassthrough(disabled bool, next http.Handler) http.Handler {
	if disabled {
		return next
	}
	return authMiddleware(next)
}

// before forwarding them to the next handler.
//
// Authentication is attempted in the following order:
//  1. Authorization: Bearer {KEY} header (takes priority)
//  2. URL path pattern /{KEY}/mcp
//
// The /health path is exempt from authentication.
// On missing or invalid authentication, returns 401 with JSON error body.
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health endpoint
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		var apiKey string

		// 1. Check Authorization: Bearer header (priority per D-01)
		auth := r.Header.Get("Authorization")
		if auth != "" && strings.HasPrefix(auth, "Bearer ") {
			apiKey = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		}

		// 2. Fall back to path-based: /{KEY}/mcp (AUTH-01, D-03)
		if apiKey == "" {
			pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
			if len(pathParts) >= 2 && pathParts[1] == "mcp" {
				apiKey = pathParts[0]
				// Strip the key segment from the path
				newPath := "/" + strings.Join(pathParts[1:], "/")
				r.URL.Path = newPath
				r.URL.RawPath = ""
			}
		}

		// 3. No API key found - reject with 401 (D-04)
		if apiKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			resp := authErrorResponse{
				Error: "Missing API key. Use path format /{API_KEY}/mcp or Authorization: Bearer {API_KEY} header",
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			}
			return
		}

		// Store API key in context and forward to next handler
		ctx := context.WithValue(r.Context(), apiKeyContextKey, apiKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
