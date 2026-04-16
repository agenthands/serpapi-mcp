// Package search provides the SerpApi search tool implementation.
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/agenthands/serpapi-mcp/internal/middleware"
	"github.com/agenthands/serpapi-mcp/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// serpapiBaseURL is the SerpApi search endpoint. Overridden in tests to point
// to a mock HTTP server.
var serpapiBaseURL = "https://serpapi.com/search"

// compactRemoveFields lists the top-level keys stripped from SerpApi responses
// when mode is "compact" (D-08).
var compactRemoveFields = []string{
	"search_metadata",
	"search_parameters",
	"search_information",
	"pagination",
	"serpapi_pagination",
}

// searchInput mirrors the tool's JSON input schema for deserialization.
type searchInput struct {
	Params map[string]any `json:"params"`
	Mode   string         `json:"mode"`
}

// RegisterSearchTool registers the "search" tool on the given MCP server.
// The tool delegates to the SerpApi HTTP API, supports complete/compact modes,
// and maps all error types to MCP-compliant IsError responses (D-01 through D-08).
func RegisterSearchTool(srv *mcp.Server, logger *slog.Logger) {
	tool := &mcp.Tool{
		Name:        "search",
		Description: "Universal search tool supporting all SerpApi engines and result types.\n\nThis tool consolidates weather, stock, and general search functionality into a single interface.\nIt processes multiple result types and returns structured JSON output.\n\nArgs:\n    params: Dictionary of engine-specific parameters. Common parameters include:\n        - q: Search query (required for most engines)\n        - engine: Search engine to use (default: \"google_light\")\n        - location: Geographic location filter\n        - num: Number of results to return\n\n    mode: Response mode (default: \"complete\")\n        - \"complete\": Returns full JSON response with all fields\n        - \"compact\": Returns JSON response with metadata fields removed\n\nReturns:\n    A JSON string containing search results or an error message.\n\nExamples:\n    Weather: {\"params\": {\"q\": \"weather in London\", \"engine\": \"google\"}, \"mode\": \"complete\"}\n    Stock: {\"params\": {\"q\": \"AAPL stock\", \"engine\": \"google\"}, \"mode\": \"complete\"}\n    General: {\"params\": {\"q\": \"coffee shops\", \"engine\": \"google_light\", \"location\": \"Austin, TX\"}, \"mode\": \"complete\"}\n    Compact: {\"params\": {\"q\": \"news\"}, \"mode\": \"compact\"}\n\nSupported engines include (not limited to):\n    - google, google_light, google_flights, google_hotels,\n    - google_images, google_news, google_local, google_shopping,\n    - google_jobs, bing, yahoo, duckduckgo, youtube_search, baidu, ebay\n\nEngine params are available via resources at serpapi://engines/<engine> (index: serpapi://engines).",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"params": {
					"type": "object",
					"description": "Dictionary of engine-specific parameters. Common: q (query), engine, location, num"
				},
				"mode": {
					"type": "string",
					"enum": ["complete", "compact"],
					"default": "complete",
					"description": "Response mode: complete returns full JSON, compact removes metadata fields"
				}
			},
			"required": ["params"]
		}`),
	}

	handler := func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return callSearchTool(ctx, req)
	}

	srv.AddTool(tool, handler)
}

// callSearchTool is the core tool handler, extracted for testability.
func callSearchTool(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Recover from any unexpected panic
	defer func() {
		if r := recover(); r != nil {
			// Already handled by returning an error result
		}
	}()

	// Extract correlation ID for structured logging (OBS-01)
	corrID := middleware.CorrelationIDFromContext(ctx)

	// 1. Unmarshal arguments
	var input searchInput
	if req.Params.Arguments != nil {
		if err := json.Unmarshal(req.Params.Arguments, &input); err != nil {
			return toolError("search_error", fmt.Sprintf("failed to parse arguments: %v", err)), nil
		}
	}

	// 2. Default mode to "complete" (D-05)
	if input.Mode == "" {
		input.Mode = "complete"
	}

	// 3. Extract API key from request context (D-06)
	apiKey := server.APIKeyFromContext(ctx)
	if apiKey == "" {
		return toolError("missing_api_key", "No API key found in request context. Use path format /{API_KEY}/mcp or Authorization: Bearer {API_KEY} header."), nil
	}

	// 4. Default engine to "google_light" (D-05, SRCH-02)
	if input.Params == nil {
		input.Params = map[string]any{}
	}
	if _, hasEngine := input.Params["engine"]; !hasEngine {
		input.Params["engine"] = "google_light"
	}

	// 4.5 Input validation — runs BEFORE any SerpApi HTTP call (VAL-01, VAL-02, VAL-03)
	engine, _ := input.Params["engine"].(string)

	if err := ValidateEngine(engine); err != nil {
		slog.Error("search validation failed", "correlation_id", corrID, "error", err, "engine", engine, "mode", input.Mode)
		return toolError("invalid_engine", err.Error()), nil
	}

	if err := ValidateMode(input.Mode); err != nil {
		slog.Error("search validation failed", "correlation_id", corrID, "error", err, "engine", engine, "mode", input.Mode)
		return toolError("invalid_mode", err.Error()), nil
	}

	if err := ValidateRequiredParams(engine, input.Params); err != nil {
		slog.Error("search validation failed", "correlation_id", corrID, "error", err, "engine", engine, "mode", input.Mode)
		return toolError("missing_params", err.Error()), nil
	}

	slog.Info("search request", "correlation_id", corrID, "engine", engine, "mode", input.Mode, "params_count", len(input.Params))

	// 5. Construct SerpApi URL
	u, err := url.Parse(serpapiBaseURL)
	if err != nil {
		return toolError("search_error", fmt.Sprintf("invalid SerpApi URL: %v", err)), nil
	}

	q := u.Query()
	q.Set("api_key", apiKey)
	for k, v := range input.Params {
		q.Set(k, fmt.Sprintf("%v", v))
	}
	u.RawQuery = q.Encode()

	// 6. Make GET request via http.Client with 30s timeout (D-07)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return toolError("search_error", fmt.Sprintf("SerpApi request failed: %v", err)), nil
	}
	defer resp.Body.Close()

	// 7. Check response status code (D-03, SRCH-07)
	switch resp.StatusCode {
	case http.StatusOK:
		// continue processing
	case http.StatusTooManyRequests:
		return toolError("rate_limited", "Rate limit exceeded. Please try again later."), nil
	case http.StatusUnauthorized:
		return toolError("invalid_api_key", "Invalid SerpApi API key. Check your API key in the path or Authorization header."), nil
	case http.StatusForbidden:
		return toolError("forbidden", "SerpApi API key forbidden. Verify your subscription and key validity."), nil
	default:
		body, _ := io.ReadAll(resp.Body)
		return toolError("search_error", fmt.Sprintf("SerpApi returned status %d: %s", resp.StatusCode, string(body))), nil
	}

	// 8. Read and parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return toolError("search_error", fmt.Sprintf("failed to read SerpApi response: %v", err)), nil
	}

	// 9. Parse JSON — handle both object and non-object responses
	var result any
	if err := json.Unmarshal(body, &result); err != nil {
		return toolError("search_error", fmt.Sprintf("failed to parse SerpApi response: %v", err)), nil
	}

	resultMap, isObject := result.(map[string]any)

	// 10. Compact mode: remove metadata fields (D-08)
	if isObject && strings.EqualFold(input.Mode, "compact") {
		for _, field := range compactRemoveFields {
			delete(resultMap, field)
		}
	}

	// 11. Marshal result back to JSON
	var resultJSON []byte
	if isObject {
		resultJSON, err = json.Marshal(resultMap)
	} else {
		resultJSON, err = json.Marshal(result)
	}
	if err != nil {
		return toolError("search_error", fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(resultJSON)}},
		IsError: false,
	}, nil
}

// toolError creates a CallToolResult with IsError=true and a flat JSON error body.
// The body format is: {"error": code, "message": message}
// This implements D-01 (tool errors use MCP IsError, not string prefixes) and D-02.
func toolError(code, message string) *mcp.CallToolResult {
	errBody := map[string]string{
		"error":   code,
		"message": message,
	}
	errJSON, _ := json.Marshal(errBody)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(errJSON)}},
		IsError: true,
	}
}
