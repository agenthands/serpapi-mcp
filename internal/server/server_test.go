package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewMCPServer(t *testing.T) {
	cfg := Config{
		Host:        "0.0.0.0",
		Port:        8000,
		CorsOrigins: "*",
	}
	srv := NewMCPServer(cfg, "test-version")
	if srv == nil {
		t.Fatal("NewMCPServer returned nil")
	}
	if srv.MCPServer == nil {
		t.Fatal("MCPServer.MCPServer is nil")
	}
}

func TestHealthcheckEndpoint(t *testing.T) {
	cfg := Config{
		Host:        "0.0.0.0",
		Port:        8000,
		CorsOrigins: "*",
	}
	srv := NewMCPServer(cfg, "test-version")

	// Build the handler chain: CORS → mux (with /mcp and /health)
	handler := srv.buildHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", contentType)
	}

	body := w.Body.String()
	if !containsJSON(body, `"status"`, `"healthy"`) {
		t.Fatalf("expected body to contain \"status\":\"healthy\", got %s", body)
	}
	if !containsJSON(body, `"service"`, `"SerpApi MCP Server"`) {
		t.Fatalf("expected body to contain \"service\":\"SerpApi MCP Server\", got %s", body)
	}
}

func TestGracefulShutdown(t *testing.T) {
	cfg := Config{
		Host:        "127.0.0.1",
		Port:        0, // will use random port
		CorsOrigins: "*",
	}
	srv := NewMCPServer(cfg, "test-version")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- srv.Run(ctx)
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context to trigger shutdown
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}

// containsJSON is a simple helper to check if a JSON body contains key-value pair.
func containsJSON(body, key, value string) bool {
	// Simple string containment check for test purposes
	return contains(body, key) && contains(body, value)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
