package search

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/agenthands/serpapi-mcp/internal/engines"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// setupEnginesForTest loads engine schemas so validation functions can access them.
// It uses the repo-root engines/ directory, which should be present in any checkout.
func setupEnginesForTest(t *testing.T) {
	t.Helper()
	// Only load if not already loaded (engines.EngineNames returns non-nil when loaded)
	if engines.EngineNames() != nil {
		return
	}
	// Create a minimal MCP server for resource registration during LoadAndRegister
	srv := mcp.NewServer(
		&mcp.Implementation{Name: "validation-test", Version: "0.0.1"},
		&mcp.ServerOptions{},
	)
	// Load from repo-root engines/ directory relative to the test working directory
	logger := slog.Default()
	_, err := engines.LoadAndRegister(srv, "../../engines", logger)
	if err != nil {
		t.Fatalf("Failed to load engine schemas for validation tests: %v", err)
	}
}

// TestValidateEngineValid tests that a known engine name returns nil error.
func TestValidateEngineValid(t *testing.T) {
	setupEnginesForTest(t)

	err := ValidateEngine("google_light")
	if err != nil {
		t.Errorf("expected nil error for valid engine 'google_light', got: %v", err)
	}
}

// TestValidateEngineInvalid tests that an invalid engine name returns an error
// containing "invalid_engine" and listing available engines.
func TestValidateEngineInvalid(t *testing.T) {
	setupEnginesForTest(t)

	err := ValidateEngine("nonexistent_engine_xyz")
	if err == nil {
		t.Fatal("expected error for invalid engine, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "invalid_engine") {
		t.Errorf("expected error to contain 'invalid_engine', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "nonexistent_engine_xyz") {
		t.Errorf("expected error to contain the invalid engine name, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Available engines:") {
		t.Errorf("expected error to list available engines, got: %s", errMsg)
	}
}

// TestValidateModeValid tests that valid modes "complete" and "compact" return nil.
func TestValidateModeValid(t *testing.T) {
	for _, mode := range []string{"complete", "compact"} {
		err := ValidateMode(mode)
		if err != nil {
			t.Errorf("expected nil error for valid mode %q, got: %v", mode, err)
		}
	}
}

// TestValidateModeInvalid tests that an invalid mode returns an error containing "invalid_mode".
func TestValidateModeInvalid(t *testing.T) {
	err := ValidateMode("detailed")
	if err == nil {
		t.Fatal("expected error for invalid mode 'detailed', got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "invalid_mode") {
		t.Errorf("expected error to contain 'invalid_mode', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "detailed") {
		t.Errorf("expected error to contain the invalid mode name, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "complete") || !strings.Contains(errMsg, "compact") {
		t.Errorf("expected error to mention valid modes, got: %s", errMsg)
	}
}

// TestValidateRequiredParamsPresent tests that no error is returned when all required params are present.
func TestValidateRequiredParamsPresent(t *testing.T) {
	setupEnginesForTest(t)

	params := map[string]any{"q": "test query"}
	err := ValidateRequiredParams("google_light", params)
	if err != nil {
		t.Errorf("expected nil error when all required params present, got: %v", err)
	}
}

// TestValidateRequiredParamsMissing tests that an error is returned listing missing params.
func TestValidateRequiredParamsMissing(t *testing.T) {
	setupEnginesForTest(t)

	params := map[string]any{"location": "Austin"}
	err := ValidateRequiredParams("google_light", params)
	if err == nil {
		t.Fatal("expected error for missing required param 'q', got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "missing_params") {
		t.Errorf("expected error to contain 'missing_params', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "q") {
		t.Errorf("expected error to list missing param 'q', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "google_light") {
		t.Errorf("expected error to mention engine name, got: %s", errMsg)
	}
}

// TestSearchValidationPreventsHTTPCall verifies that validation errors are returned
// before any SerpApi HTTP request is made (no HTTP call on validation failure).
func TestSearchValidationPreventsHTTPCall(t *testing.T) {
	// Set up a mock server that would fail if contacted
	mockSerpAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("SerpApi HTTP request should not be made when validation fails")
	}))
	defer mockSerpAPI.Close()

	origResolver := serpapiBaseURLResolver
	serpapiBaseURLResolver = func() string { return mockSerpAPI.URL }
	defer func() { serpapiBaseURLResolver = origResolver }()

	// Also ensure engines are loaded for validation
	setupEnginesForTest(t)

	// Test with invalid engine
	args := map[string]any{
		"params": map[string]any{"q": "test", "engine": "nonexistent_engine_xyz"},
	}
	argsJSON, _ := json.Marshal(args)

	result, err := callWithAPIKey(argsJSON, "test-api-key")
	if err != nil {
		t.Fatalf("search tool returned error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected IsError=true for invalid engine")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	if !strings.Contains(textContent.Text, "invalid_engine") {
		t.Errorf("expected error code 'invalid_engine' in response, got: %s", textContent.Text)
	}
}

// TestValidateEngineEmpty verifies that an empty engine name returns an
// error containing "invalid_engine".
func TestValidateEngineEmpty(t *testing.T) {
	setupEnginesForTest(t)

	err := ValidateEngine("")
	if err == nil {
		t.Fatal("expected error for empty engine name, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "invalid_engine") {
		t.Errorf("expected error to contain 'invalid_engine', got: %s", errMsg)
	}
}

// TestValidateModeEmpty verifies that an empty mode returns an error
// containing "invalid_mode".
func TestValidateModeEmpty(t *testing.T) {
	err := ValidateMode("")
	if err == nil {
		t.Fatal("expected error for empty mode, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "invalid_mode") {
		t.Errorf("expected error to contain 'invalid_mode', got: %s", errMsg)
	}
}

// TestValidateModeCaseSensitive verifies that mode validation is case-sensitive:
// "COMPLETE" and "Compact" should be rejected (only lowercase accepted).
func TestValidateModeCaseSensitive(t *testing.T) {
	for _, mode := range []string{"COMPLETE", "Compact", "COMPACT"} {
		err := ValidateMode(mode)
		if err == nil {
			t.Errorf("expected error for uppercase mode %q, got nil", mode)
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "invalid_mode") {
			t.Errorf("expected 'invalid_mode' for mode %q, got: %s", mode, errMsg)
		}
	}
}

// TestValidateRequiredParamsUnknownEngine verifies that an unknown engine name
// causes RequiredParams to return nil, so ValidateRequiredParams skips validation
// (ValidateEngine catches invalid names separately).
func TestValidateRequiredParamsUnknownEngine(t *testing.T) {
	setupEnginesForTest(t)

	err := ValidateRequiredParams("nonexistent_engine_xyz", map[string]any{"q": "test"})
	if err != nil {
		t.Errorf("expected nil error for unknown engine (validation skipped), got: %v", err)
	}
}

// TestValidateRequiredParamsNilParams verifies that nil params map doesn't crash
// the validator. It should report missing required parameters.
func TestValidateRequiredParamsNilParams(t *testing.T) {
	setupEnginesForTest(t)

	// google_light requires "q" — nil map access in Go returns zero value safely
	err := ValidateRequiredParams("google_light", nil)
	if err == nil {
		t.Fatal("expected error for nil params with required fields, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "missing_params") {
		t.Errorf("expected 'missing_params' error, got: %s", errMsg)
	}
}
