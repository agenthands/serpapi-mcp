package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestBearerHeaderAuth verifies that a request with Authorization: Bearer {KEY}
// passes through with the API key stored in context.
func TestBearerHeaderAuth(t *testing.T) {
	var gotKey string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = APIKeyFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := authMiddleware(inner)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer testkey123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	if gotKey != "testkey123" {
		t.Fatalf("expected API key 'testkey123' in context, got %q", gotKey)
	}
}

// TestBearerHeaderAuthWithPrefix verifies extraction of keys with dashes.
func TestBearerHeaderAuthWithPrefix(t *testing.T) {
	var gotKey string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = APIKeyFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := authMiddleware(inner)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer my-api-key-456")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if gotKey != "my-api-key-456" {
		t.Fatalf("expected API key 'my-api-key-456', got %q", gotKey)
	}
}

// TestPathBasedAuth verifies that a request to /{KEY}/mcp extracts the key
// and rewrites the path to /mcp.
func TestPathBasedAuth(t *testing.T) {
	var gotKey string
	var gotPath string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = APIKeyFromContext(r.Context())
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	handler := authMiddleware(inner)

	req := httptest.NewRequest(http.MethodPost, "/testkey123/mcp", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
	if gotKey != "testkey123" {
		t.Fatalf("expected API key 'testkey123' in context, got %q", gotKey)
	}
	if gotPath != "/mcp" {
		t.Fatalf("expected path rewritten to '/mcp', got %q", gotPath)
	}
}

// TestPathBasedAuthStripsKey verifies that the key segment is stripped from the path.
func TestPathBasedAuthStripsKey(t *testing.T) {
	var gotPath string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	handler := authMiddleware(inner)

	req := httptest.NewRequest(http.MethodPost, "/somekey/mcp", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if gotPath != "/mcp" {
		t.Fatalf("expected path '/mcp', got %q", gotPath)
	}
}

// TestMissingAPIKey verifies that a request to /mcp without auth
// returns 401 with JSON error body.
func TestMissingAPIKey(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called for unauthenticated request")
	})

	handler := authMiddleware(inner)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Fatalf("expected Content-Type application/json, got %s", contentType)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON body: %v", err)
	}
	if !strings.Contains(body["error"], "Missing API key") {
		t.Fatalf("expected error containing 'Missing API key', got %q", body["error"])
	}
}

// TestHealthExempt verifies that /health passes through without auth.
func TestHealthExempt(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := authMiddleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Fatal("expected inner handler to be called for /health")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	// API key should be empty since we didn't provide auth, but request passed through
	key := APIKeyFromContext(req.Context())
	if key != "" {
		t.Fatalf("expected no API key in context for /health, got %q", key)
	}
}

// TestBearerHeaderPriority verifies that Bearer header takes priority
// over path-based key when both are present.
func TestBearerHeaderPriority(t *testing.T) {
	var gotKey string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = APIKeyFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := authMiddleware(inner)

	req := httptest.NewRequest(http.MethodPost, "/pathkey/mcp", nil)
	req.Header.Set("Authorization", "Bearer headerkey")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if gotKey != "headerkey" {
		t.Fatalf("expected Bearer header key 'headerkey' to take priority, got %q", gotKey)
	}
}

// TestInvalidBearerFormat verifies that "Basic abc123" (not Bearer) returns 401.
func TestInvalidBearerFormat(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called for invalid auth format")
	})

	handler := authMiddleware(inner)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid Bearer format, got %d", w.Code)
	}
}

// TestAPIKeyFromContext verifies the context helper returns empty string for nil/empty context.
func TestAPIKeyFromContext(t *testing.T) {
	key := APIKeyFromContext(context.Background())
	if key != "" {
		t.Fatalf("expected empty string from background context, got %q", key)
	}
}
