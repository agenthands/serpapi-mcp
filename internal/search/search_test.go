package search

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/agenthands/serpapi-mcp/internal/engines"
	"github.com/agenthands/serpapi-mcp/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// setupTestServer creates a mock HTTP server and MCP server for testing.
// The mock server responds based on query parameters:
//   - ?_test_status=429 → 429 response
//   - ?_test_status=401 → 401 response
//   - ?_test_status=403 → 403 response
//   - ?_test_status=500 → 500 response
//   - default → 200 with a JSON body including metadata fields
func setupTestServer(t *testing.T) (*httptest.Server, *mcp.Server) {
	t.Helper()

	mockSerpAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch q.Get("_test_status") {
		case "429":
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Rate limit exceeded"}`))
		case "401":
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"Invalid API key"}`))
		case "403":
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"Forbidden"}`))
		case "500":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Internal server error"}`))
		default:
			// Return a minimal valid SerpApi response with all metadata fields
			resp := map[string]any{
				"search_metadata":    map[string]any{"id": "test", "status": "Success"},
				"search_parameters":  map[string]any{"engine": "google_light", "q": "test"},
				"search_information": map[string]any{"total_results": 100},
				"pagination":         map[string]any{"current": 1, "next": "page2"},
				"serpapi_pagination": map[string]any{"next": "https://serpapi.com/page2"},
				"organic_results":    []map[string]any{{"title": "Test Result", "link": "https://example.com"}},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	}))

	origResolver := serpapiBaseURLResolver
	serpapiBaseURLResolver = func() string { return mockSerpAPI.URL }

	// Load engine schemas for validation (only if not already loaded)
	if engines.EngineNames() == nil {
		engSrv := mcp.NewServer(
			&mcp.Implementation{Name: "search-test", Version: "0.0.1"},
			&mcp.ServerOptions{},
		)
		if _, err := engines.LoadAndRegister(engSrv, "../../engines", slog.Default()); err != nil {
			t.Fatalf("Failed to load engine schemas: %v", err)
		}
	}

	mcpSrv := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "0.1.0"},
		&mcp.ServerOptions{},
	)

	t.Cleanup(func() {
		mockSerpAPI.Close()
		serpapiBaseURLResolver = origResolver
	})

	return mockSerpAPI, mcpSrv
}

// newCallToolRequest creates a CallToolRequest with the given arguments JSON and context.
func newCallToolRequest(argsJSON []byte, ctx context.Context) *mcp.CallToolRequest {
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "search",
			Arguments: json.RawMessage(argsJSON),
		},
	}
	_ = ctx // context is passed as first arg to handler, not stored in request
	return req
}

// callWithAPIKey calls callSearchTool with a context containing the given API key.
func callWithAPIKey(argsJSON []byte, apiKey string) (*mcp.CallToolResult, error) {
	ctx := server.ContextWithAPIKey(context.Background(), apiKey)
	req := newCallToolRequest(argsJSON, ctx)
	return callSearchTool(ctx, req)
}

// callWithoutAPIKey calls callSearchTool with an empty context (no API key).
func callWithoutAPIKey(argsJSON []byte) (*mcp.CallToolResult, error) {
	req := newCallToolRequest(argsJSON, context.Background())
	return callSearchTool(context.Background(), req)
}

// Test 1: Search with default engine and complete mode returns full JSON response.
func TestSearchCompleteMode(t *testing.T) {
	_, _ = setupTestServer(t)

	args := map[string]any{
		"params": map[string]any{"q": "test query"},
		"mode":   "complete",
	}
	argsJSON, _ := json.Marshal(args)

	result, err := callWithAPIKey(argsJSON, "test-api-key")
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &body); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Complete mode should include all metadata fields
	for _, field := range []string{"search_metadata", "search_parameters", "search_information", "pagination", "serpapi_pagination"} {
		if _, exists := body[field]; !exists {
			t.Errorf("complete mode: expected field %q in response", field)
		}
	}
}

// Test 2: Search with compact mode removes the 5 metadata fields.
func TestSearchCompactMode(t *testing.T) {
	_, _ = setupTestServer(t)

	args := map[string]any{
		"params": map[string]any{"q": "test query"},
		"mode":   "compact",
	}
	argsJSON, _ := json.Marshal(args)

	result, err := callWithAPIKey(argsJSON, "test-api-key")
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &body); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Compact mode should NOT include these fields
	for _, field := range []string{"search_metadata", "search_parameters", "search_information", "pagination", "serpapi_pagination"} {
		if _, exists := body[field]; exists {
			t.Errorf("compact mode: expected field %q to be removed from response", field)
		}
	}

	// But should still contain actual results
	if _, exists := body["organic_results"]; !exists {
		t.Error("compact mode: expected organic_results to remain in response")
	}
}

// Test 3: SerpApi 429 returns {IsError: true, Content contains "rate_limited"}.
func TestSearchRateLimited(t *testing.T) {
	_, _ = setupTestServer(t)

	args := map[string]any{
		"params": map[string]any{"q": "test query", "_test_status": "429"},
	}
	argsJSON, _ := json.Marshal(args)

	result, err := callWithAPIKey(argsJSON, "test-api-key")
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected IsError=true for 429 response")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "rate_limited") {
		t.Errorf("expected error code 'rate_limited' in response, got: %s", textContent.Text)
	}
}

// Test 4: SerpApi 401 returns {IsError: true, Content contains "invalid_api_key"}.
func TestSearchInvalidAPIKey(t *testing.T) {
	_, _ = setupTestServer(t)

	args := map[string]any{
		"params": map[string]any{"q": "test query", "_test_status": "401"},
	}
	argsJSON, _ := json.Marshal(args)

	result, err := callWithAPIKey(argsJSON, "bad-api-key")
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected IsError=true for 401 response")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "invalid_api_key") {
		t.Errorf("expected error code 'invalid_api_key' in response, got: %s", textContent.Text)
	}
}

// Test 5: SerpApi 403 returns {IsError: true, Content contains "forbidden"}.
func TestSearchForbidden(t *testing.T) {
	_, _ = setupTestServer(t)

	args := map[string]any{
		"params": map[string]any{"q": "test query", "_test_status": "403"},
	}
	argsJSON, _ := json.Marshal(args)

	result, err := callWithAPIKey(argsJSON, "forbidden-key")
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected IsError=true for 403 response")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "forbidden") {
		t.Errorf("expected error code 'forbidden' in response, got: %s", textContent.Text)
	}
}

// Test 6: Network/5xx error returns {IsError: true, Content contains "search_error"}.
func TestSearchServerError(t *testing.T) {
	_, _ = setupTestServer(t)

	args := map[string]any{
		"params": map[string]any{"q": "test query", "_test_status": "500"},
	}
	argsJSON, _ := json.Marshal(args)

	result, err := callWithAPIKey(argsJSON, "test-api-key")
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected IsError=true for 500 response")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "search_error") {
		t.Errorf("expected error code 'search_error' in response, got: %s", textContent.Text)
	}
}

// Test 7: Missing API key in context returns {IsError: true, Content contains "missing_api_key"}.
func TestSearchMissingAPIKey(t *testing.T) {
	_, _ = setupTestServer(t)

	args := map[string]any{
		"params": map[string]any{"q": "test query"},
	}
	argsJSON, _ := json.Marshal(args)

	result, err := callWithoutAPIKey(argsJSON)
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected IsError=true for missing API key")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "missing_api_key") {
		t.Errorf("expected error code 'missing_api_key' in response, got: %s", textContent.Text)
	}
}

// TestSearchDefaultEngine verifies that the default engine is google_light.
func TestSearchDefaultEngine(t *testing.T) {
	var receivedEngine string

	mockSerpAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedEngine = r.URL.Query().Get("engine")
		resp := map[string]any{"organic_results": []any{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockSerpAPI.Close()

	origResolver := serpapiBaseURLResolver
	serpapiBaseURLResolver = func() string { return mockSerpAPI.URL }
	defer func() { serpapiBaseURLResolver = origResolver }()

	args := map[string]any{
		"params": map[string]any{"q": "test query"},
	}
	argsJSON, _ := json.Marshal(args)

	result, err := callWithAPIKey(argsJSON, "test-api-key")
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	if receivedEngine != "google_light" {
		t.Errorf("expected default engine 'google_light', got %q", receivedEngine)
	}
}

// TestSearchMalformedJSON verifies that truncated/invalid JSON arguments
// return IsError=true with "search_error" code instead of crashing.
func TestSearchMalformedJSON(t *testing.T) {
	_, _ = setupTestServer(t)

	// Truncated JSON — missing closing brace
	argsJSON := json.RawMessage(`{"params":`)

	result, err := callWithAPIKey(argsJSON, "test-api-key")
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected IsError=true for malformed JSON input")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "search_error") {
		t.Errorf("expected error code 'search_error' in response, got: %s", textContent.Text)
	}
}

// TestSearchEmptyParams verifies that omitting the "engine" key defaults to google_light.
func TestSearchEmptyParams(t *testing.T) {
	var receivedEngine string

	mockSerpAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedEngine = r.URL.Query().Get("engine")
		resp := map[string]any{"organic_results": []any{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockSerpAPI.Close()

	origResolver := serpapiBaseURLResolver
	serpapiBaseURLResolver = func() string { return mockSerpAPI.URL }
	defer func() { serpapiBaseURLResolver = origResolver }()

	// Include "q" (required param for google_light) but omit "engine"
	args := map[string]any{
		"params": map[string]any{"q": "test query"},
	}
	argsJSON, _ := json.Marshal(args)

	result, err := callWithAPIKey(argsJSON, "test-api-key")
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	if receivedEngine != "google_light" {
		t.Errorf("expected default engine 'google_light', got %q", receivedEngine)
	}
}

// TestSearchNullArguments verifies that nil Arguments do not cause a crash.
// The handler should treat it as empty input and return an error (missing_api_key
// if no API key, or validation error if API key is provided).
func TestSearchNullArguments(t *testing.T) {
	_, _ = setupTestServer(t)

	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "search",
			Arguments: nil,
		},
	}

	// Call without API key — should return missing_api_key, not panic
	result, err := callSearchTool(context.Background(), req)
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected IsError=true for nil arguments without API key")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	// Should get missing_api_key (no API key in context) or a validation error
	if !strings.Contains(textContent.Text, "missing_api_key") && !strings.Contains(textContent.Text, "search_error") {
		t.Errorf("expected 'missing_api_key' or 'search_error' in response, got: %s", textContent.Text)
	}
}

// TestSearchCompactModeOnArrayResponse verifies that compact mode handles
// non-object responses (e.g., JSON arrays) correctly — no field removal,
// array returned unchanged.
func TestSearchCompactModeOnArrayResponse(t *testing.T) {
	mockSerpAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[1,2,3]`))
	}))
	defer mockSerpAPI.Close()

	origResolver := serpapiBaseURLResolver
	serpapiBaseURLResolver = func() string { return mockSerpAPI.URL }
	defer func() { serpapiBaseURLResolver = origResolver }()

	// Ensure engines are loaded for validation
	if engines.EngineNames() == nil {
		engSrv := mcp.NewServer(
			&mcp.Implementation{Name: "compact-array-test", Version: "0.0.1"},
			&mcp.ServerOptions{},
		)
		if _, err := engines.LoadAndRegister(engSrv, "../../engines", slog.Default()); err != nil {
			t.Fatalf("Failed to load engine schemas: %v", err)
		}
	}

	args := map[string]any{
		"params": map[string]any{"q": "test query"},
		"mode":   "compact",
	}
	argsJSON, _ := json.Marshal(args)

	result, err := callWithAPIKey(argsJSON, "test-api-key")
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	// Array responses should pass through unchanged in compact mode
	expected := `[1,2,3]`
	if strings.TrimSpace(textContent.Text) != expected {
		t.Errorf("expected array response %q, got: %s", expected, textContent.Text)
	}
}

// TestSearchCompactModeAllFieldsRemoved verifies that each of the 5
// compactRemoveFields is individually absent in compact mode.
func TestSearchCompactModeAllFieldsRemoved(t *testing.T) {
	_, _ = setupTestServer(t)

	args := map[string]any{
		"params": map[string]any{"q": "test query"},
		"mode":   "compact",
	}
	argsJSON, _ := json.Marshal(args)

	result, err := callWithAPIKey(argsJSON, "test-api-key")
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if result.IsError {
		t.Fatalf("expected success, got error: %v", result.Content)
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(textContent.Text), &body); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	// Verify each compactRemoveField is absent
	for _, field := range compactRemoveFields {
		if _, exists := body[field]; exists {
			t.Errorf("compact mode: expected field %q to be removed, but it was present", field)
		}
	}
}
