package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agenthands/serpapi-mcp/internal/engines"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

func TestServerStartsWithEngines(t *testing.T) {
	// Create a temp directory with test engine files
	dir := t.TempDir()
	engineContent := `{"engine":"test_engine","params":{"q":{"required":true,"description":"test query"}}}`
	if err := os.WriteFile(filepath.Join(dir, "test_engine.json"), []byte(engineContent), 0644); err != nil {
		t.Fatalf("failed to create test engine file: %v", err)
	}

	cfg := Config{
		Host:        "127.0.0.1",
		Port:        0,
		CorsOrigins: "*",
	}
	srv := NewMCPServer(cfg, "test-version")

	count, err := engines.LoadAndRegister(srv.MCPServer, dir, slog.Default())
	if err != nil {
		t.Fatalf("LoadAndRegister returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 engine, got %d", count)
	}
	srv.SetEngineCount(count)

	// Start the server and verify it starts up with engines loaded
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

func TestMissingEnginesDirFailsStartup(t *testing.T) {
	cfg := Config{
		Host:        "0.0.0.0",
		Port:        8000,
		CorsOrigins: "*",
	}
	srv := NewMCPServer(cfg, "test-version")

	_, err := engines.LoadAndRegister(srv.MCPServer, "/nonexistent/path/engines", slog.Default())
	if err == nil {
		t.Fatal("expected error for non-existent engines directory")
	}
}

// TestHealthcheckJSONStructure verifies the exact JSON structure returned by /health:
// {"status":"healthy","service":"SerpApi MCP Server"} with Content-Type application/json.
func TestHealthcheckJSONStructure(t *testing.T) {
	cfg := Config{
		Host:        "0.0.0.0",
		Port:        8000,
		CorsOrigins: "*",
	}
	srv := NewMCPServer(cfg, "test-version")
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

	var body healthResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON body: %v", err)
	}
	if body.Status != "healthy" {
		t.Fatalf("expected status 'healthy', got %q", body.Status)
	}
	if body.Service != "SerpApi MCP Server" {
		t.Fatalf("expected service 'SerpApi MCP Server', got %q", body.Service)
	}
}

// TestIntegrationAuthDisabledPassThrough verifies that Config{AuthDisabled: true}
// allows requests to /mcp without authentication — auth middleware is skipped.
func TestIntegrationAuthDisabledPassThrough(t *testing.T) {
	cfg := Config{
		Host:         "0.0.0.0",
		Port:         8000,
		CorsOrigins:  "*",
		AuthDisabled: true,
	}
	srv := NewMCPServer(cfg, "test-version")
	handler := srv.buildHandler()

	// POST /mcp without auth — auth disabled, so it should reach MCP handler
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Auth disabled: should NOT return 401. MCP handler may return 4xx for
	// bad request, but it's not an auth failure.
	if w.Code == http.StatusUnauthorized {
		t.Fatalf("expected auth to be disabled (not 401), got 401")
	}
}

// TestIntegrationCorrelationIDInResponse verifies that the X-Correlation-ID
// header is present in healthcheck responses (correlation middleware is active).
func TestIntegrationCorrelationIDInResponse(t *testing.T) {
	cfg := Config{
		Host:        "0.0.0.0",
		Port:        8000,
		CorsOrigins: "*",
	}
	srv := NewMCPServer(cfg, "test-version")
	handler := srv.buildHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	correlationID := w.Header().Get("X-Correlation-ID")
	if correlationID == "" {
		t.Fatal("expected X-Correlation-ID header in response, got empty string")
	}
	// Correlation IDs are 32-char hex strings (16 bytes encoded as hex)
	if len(correlationID) != 32 {
		t.Fatalf("expected 32-char correlation ID, got %d chars: %q", len(correlationID), correlationID)
	}
	for _, c := range correlationID {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("expected hex chars in correlation ID, got %q in %q", c, correlationID)
		}
	}
}

func TestEngineResourceReadViaMCP(t *testing.T) {
	// Create a temp directory with test engine files
	dir := t.TempDir()
	engineContent1 := `{"engine":"alpha","params":{"q":{"required":true,"description":"query"}}}`
	engineContent2 := `{"engine":"bravo","params":{"q":{"required":true,"description":"query"}}}`
	if err := os.WriteFile(filepath.Join(dir, "alpha.json"), []byte(engineContent1), 0644); err != nil {
		t.Fatalf("failed to create test engine file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bravo.json"), []byte(engineContent2), 0644); err != nil {
		t.Fatalf("failed to create test engine file: %v", err)
	}

	// Build MCP server and load engines
	mcpSrv := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "test"},
		&mcp.ServerOptions{Logger: slog.Default()},
	)

	count, err := engines.LoadAndRegister(mcpSrv, dir, slog.Default())
	if err != nil {
		t.Fatalf("LoadAndRegister returned error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 engines, got %d", count)
	}

	// Connect a client to test resource reads
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := mcpSrv.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect failed: %v", err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	cs, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect failed: %v", err)
	}

	// Read serpapi://engines index resource
	indexResult, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "serpapi://engines"})
	if err != nil {
		t.Fatalf("ReadResource for engines index failed: %v", err)
	}
	if len(indexResult.Contents) == 0 {
		t.Fatal("expected content in engines index")
	}

	var index map[string]any
	if err := json.Unmarshal([]byte(indexResult.Contents[0].Text), &index); err != nil {
		t.Fatalf("failed to parse engines index: %v", err)
	}
	if int(index["count"].(float64)) != 2 {
		t.Fatalf("expected count 2, got %v", index["count"])
	}

	// Read per-engine resource
	engineResult, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "serpapi://engines/alpha"})
	if err != nil {
		t.Fatalf("ReadResource for alpha engine failed: %v", err)
	}
	if len(engineResult.Contents) == 0 {
		t.Fatal("expected content in engine resource")
	}

	var engineSchema map[string]any
	if err := json.Unmarshal([]byte(engineResult.Contents[0].Text), &engineSchema); err != nil {
		t.Fatalf("failed to parse engine schema: %v", err)
	}
	if engineSchema["engine"] != "alpha" {
		t.Fatalf("expected engine 'alpha', got %v", engineSchema["engine"])
	}
}
