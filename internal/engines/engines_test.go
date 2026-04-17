package engines

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createTestEngineFile creates a temporary engine JSON file in the given directory.
func createTestEngineFile(t *testing.T, dir, engineName, content string) {
	t.Helper()
	if content == "" {
		content = `{"engine":"` + engineName + `","params":{"q":{"required":true,"description":"test query"}}}`
	}
	err := os.WriteFile(filepath.Join(dir, engineName+".json"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create test engine file: %v", err)
	}
}

// setupTestServer creates an MCP server for testing and returns it with a cleanup function.
func setupTestServer(t *testing.T) (*mcp.Server, *mcp.ClientSession, context.Context) {
	t.Helper()
	srv := mcp.NewServer(
		&mcp.Implementation{Name: "test-server", Version: "test"},
		&mcp.ServerOptions{Logger: slog.Default()},
	)

	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect failed: %v", err)
	}

	client := mcp.NewClient(
		&mcp.Implementation{Name: "test-client", Version: "test"},
		nil,
	)
	cs, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect failed: %v", err)
	}

	return srv, cs, ctx
}

func TestLoadAndRegister_ValidEngines(t *testing.T) {
	dir := t.TempDir()
	createTestEngineFile(t, dir, "google_light", "")
	createTestEngineFile(t, dir, "bing", "")

	srv, cs, ctx := setupTestServer(t)

	logger := slog.Default()
	count, err := LoadAndRegister(srv, dir, logger)
	if err != nil {
		t.Fatalf("LoadAndRegister returned error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 engines, got %d", count)
	}

	// Verify we can list resources
	result, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}

	// Should have: serpapi://engines + 2 per-engine resources = 3 total
	if len(result.Resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(result.Resources))
	}

	// Verify the index resource exists
	var foundIndex bool
	for _, r := range result.Resources {
		if r.URI == "serpapi://engines" {
			foundIndex = true
			break
		}
	}
	if !foundIndex {
		t.Fatal("serpapi://engines index resource not found")
	}
}

func TestLoadAndRegister_MissingDir(t *testing.T) {
	srv, _, _ := setupTestServer(t)

	logger := slog.Default()
	count, err := LoadAndRegister(srv, "/nonexistent/path/engines", logger)
	if err == nil {
		t.Fatal("expected error for missing directory")
	}
	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
}

func TestLoadAndRegister_CorruptJSON(t *testing.T) {
	dir := t.TempDir()
	createTestEngineFile(t, dir, "good_engine", "")
	createTestEngineFile(t, dir, "bad_engine", `{this is not valid json`)

	srv, _, _ := setupTestServer(t)

	logger := slog.Default()
	count, err := LoadAndRegister(srv, dir, logger)
	if err == nil {
		t.Fatal("expected error for corrupt JSON")
	}
	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
}

func TestLoadAndRegister_InvalidName(t *testing.T) {
	dir := t.TempDir()
	createTestEngineFile(t, dir, "valid_engine", "")
	// Create files with invalid names (hyphens, uppercase)
	createTestEngineFile(t, dir, "google-light", `{"engine":"google-light","params":{}}`)
	createTestEngineFile(t, dir, "Google", `{"engine":"Google","params":{}}`)

	srv, cs, ctx := setupTestServer(t)

	logger := slog.Default()
	count, err := LoadAndRegister(srv, dir, logger)
	if err != nil {
		t.Fatalf("LoadAndRegister returned error: %v", err)
	}
	// Only valid_engine should be loaded, hyphens and uppercase should be skipped
	if count != 1 {
		t.Fatalf("expected 1 engine (skipped invalid names), got %d", count)
	}

	// Verify the valid engine's per-engine resource exists
	result, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}

	// Should have: serpapi://engines + 1 per-engine = 2 total
	if len(result.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(result.Resources))
	}
}

func TestEnginesIndexResource(t *testing.T) {
	dir := t.TempDir()
	createTestEngineFile(t, dir, "bing", "")
	createTestEngineFile(t, dir, "google_light", "")

	srv, cs, ctx := setupTestServer(t)

	logger := slog.Default()
	_, err := LoadAndRegister(srv, dir, logger)
	if err != nil {
		t.Fatalf("LoadAndRegister returned error: %v", err)
	}

	// Read the engines index resource
	readResult, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "serpapi://engines"})
	if err != nil {
		t.Fatalf("ReadResource failed: %v", err)
	}

	if len(readResult.Contents) == 0 {
		t.Fatal("expected at least one content item")
	}

	content := readResult.Contents[0]
	if content.MIMEType != "application/json" {
		t.Fatalf("expected MIME type application/json, got %s", content.MIMEType)
	}

	// Parse the index JSON
	var index map[string]any
	if err := json.Unmarshal([]byte(content.Text), &index); err != nil {
		t.Fatalf("failed to parse engines index JSON: %v", err)
	}

	// Verify count
	count, ok := index["count"].(float64)
	if !ok {
		t.Fatalf("expected count to be a number, got %T", index["count"])
	}
	if int(count) != 2 {
		t.Fatalf("expected count 2, got %d", int(count))
	}

	// Verify engines list (should be sorted)
	engines, ok := index["engines"].([]any)
	if !ok {
		t.Fatalf("expected engines to be an array, got %T", index["engines"])
	}
	if len(engines) != 2 {
		t.Fatalf("expected 2 engines in list, got %d", len(engines))
	}
	// Sorted order: bing, google_light
	if engines[0] != "bing" {
		t.Fatalf("expected first engine to be 'bing', got %v", engines[0])
	}
	if engines[1] != "google_light" {
		t.Fatalf("expected second engine to be 'google_light', got %v", engines[1])
	}

	// Verify resources list
	resources, ok := index["resources"].([]any)
	if !ok {
		t.Fatalf("expected resources to be an array, got %T", index["resources"])
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 resource URIs, got %d", len(resources))
	}
	if resources[0] != "serpapi://engines/bing" {
		t.Fatalf("expected first resource URI 'serpapi://engines/bing', got %v", resources[0])
	}
	if resources[1] != "serpapi://engines/google_light" {
		t.Fatalf("expected second resource URI 'serpapi://engines/google_light', got %v", resources[1])
	}

	// Verify schema info exists
	schemaInfo, ok := index["schema"].(map[string]any)
	if !ok {
		t.Fatalf("expected schema to be an object, got %T", index["schema"])
	}
	if schemaInfo["params_key"] != "params" {
		t.Fatalf("expected params_key 'params', got %v", schemaInfo["params_key"])
	}
	if schemaInfo["common_params_key"] != "common_params" {
		t.Fatalf("expected common_params_key 'common_params', got %v", schemaInfo["common_params_key"])
	}
}

func TestEngineSchemaResource(t *testing.T) {
	dir := t.TempDir()
	engineContent := `{"engine":"google_light","params":{"q":{"required":true,"description":"test query"}}}`
	createTestEngineFile(t, dir, "google_light", engineContent)

	srv, cs, ctx := setupTestServer(t)

	logger := slog.Default()
	_, err := LoadAndRegister(srv, dir, logger)
	if err != nil {
		t.Fatalf("LoadAndRegister returned error: %v", err)
	}

	// Read the per-engine resource
	readResult, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "serpapi://engines/google_light"})
	if err != nil {
		t.Fatalf("ReadResource failed: %v", err)
	}

	if len(readResult.Contents) == 0 {
		t.Fatal("expected at least one content item")
	}

	content := readResult.Contents[0]
	if content.MIMEType != "application/json" {
		t.Fatalf("expected MIME type application/json, got %s", content.MIMEType)
	}
	if content.URI != "serpapi://engines/google_light" {
		t.Fatalf("expected URI serpapi://engines/google_light, got %s", content.URI)
	}

	// Parse the engine JSON
	var schema map[string]any
	if err := json.Unmarshal([]byte(content.Text), &schema); err != nil {
		t.Fatalf("failed to parse engine schema JSON: %v", err)
	}

	if schema["engine"] != "google_light" {
		t.Fatalf("expected engine 'google_light', got %v", schema["engine"])
	}

	params, ok := schema["params"].(map[string]any)
	if !ok {
		t.Fatalf("expected params to be an object, got %T", schema["params"])
	}
	if _, ok := params["q"]; !ok {
		t.Fatal("expected params to contain 'q'")
	}
}

func TestLoadAndRegister_EngineFieldMismatch(t *testing.T) {
	dir := t.TempDir()
	// JSON file named "google_light.json" but engine field says "google"
	createTestEngineFile(t, dir, "google_light", `{"engine":"google","params":{}}`)

	srv, _, _ := setupTestServer(t)

	logger := slog.Default()
	_, err := LoadAndRegister(srv, dir, logger)
	if err == nil {
		t.Fatal("expected error for engine field mismatch")
	}
}

func TestLoadAndRegister_EmptyDir(t *testing.T) {
	dir := t.TempDir() // empty temp dir

	srv, cs, ctx := setupTestServer(t)

	logger := slog.Default()
	count, err := LoadAndRegister(srv, dir, logger)
	if err != nil {
		t.Fatalf("LoadAndRegister returned error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 engines for empty dir, got %d", count)
	}

	// Should still have the serpapi://engines index resource with count=0
	result, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}
	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource (index only), got %d", len(result.Resources))
	}
}

// TestLoadAndRegister_RealEnginesDir verifies all 107 engine JSON files
// from the project's engines/ directory can be loaded and registered.
// This test is skipped if the engines/ directory is not found (e.g., in CI
// with different working directory).
func TestLoadAndRegister_RealEnginesDir(t *testing.T) {
	// Resolve the engines directory relative to this test file
	enginesDir := filepath.Join("..", "..", "engines")
	if _, err := os.Stat(enginesDir); os.IsNotExist(err) {
		t.Skip("skipping: engines/ directory not found (expected at repo root)")
	}

	srv, cs, ctx := setupTestServer(t)

	logger := slog.Default()
	count, err := LoadAndRegister(srv, enginesDir, logger)
	if err != nil {
		t.Fatalf("LoadAndRegister with real engines dir returned error: %v", err)
	}
	if count != 107 {
		t.Fatalf("expected 107 engines, got %d", count)
	}

	// Verify we can list all 108 resources (107 engines + index)
	result, err := cs.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}
	if len(result.Resources) != 108 {
		t.Fatalf("expected 108 resources (1 index + 107 engines), got %d", len(result.Resources))
	}

	// Read the index and verify count
	indexResult, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "serpapi://engines"})
	if err != nil {
		t.Fatalf("ReadResource for engines index failed: %v", err)
	}

	var index map[string]any
	if err := json.Unmarshal([]byte(indexResult.Contents[0].Text), &index); err != nil {
		t.Fatalf("failed to parse engines index: %v", err)
	}
	if int(index["count"].(float64)) != 107 {
		t.Fatalf("expected count 107 in index, got %v", index["count"])
	}

	// Read one per-engine resource (google_light)
	engineResult, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "serpapi://engines/google_light"})
	if err != nil {
		t.Fatalf("ReadResource for google_light failed: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal([]byte(engineResult.Contents[0].Text), &schema); err != nil {
		t.Fatalf("failed to parse google_light schema: %v", err)
	}
	if schema["engine"] != "google_light" {
		t.Fatalf("expected engine 'google_light', got %v", schema["engine"])
	}
}

// TestRequiredParamsForEngineWithRequired verifies that RequiredParams returns
// the correct list of required parameter names for an engine with required params.
func TestRequiredParamsForEngineWithRequired(t *testing.T) {
	dir := t.TempDir()
	content := `{"engine":"google_light","params":{"q":{"required":true,"description":"query"},"hl":{"required":false,"description":"language"}}}`
	createTestEngineFile(t, dir, "google_light", content)

	srv, _, _ := setupTestServer(t)

	logger := slog.Default()
	_, err := LoadAndRegister(srv, dir, logger)
	if err != nil {
		t.Fatalf("LoadAndRegister returned error: %v", err)
	}

	required := RequiredParams("google_light")
	if len(required) != 1 || required[0] != "q" {
		t.Errorf("expected RequiredParams to return [\"q\"], got: %v", required)
	}
}

// TestRequiredParamsForEngineWithNoRequired verifies that RequiredParams returns
// nil or empty for an engine with no required params.
func TestRequiredParamsForEngineWithNoRequired(t *testing.T) {
	dir := t.TempDir()
	content := `{"engine":"sandbox","params":{"q":{"required":false,"description":"optional query"}}}`
	createTestEngineFile(t, dir, "sandbox", content)

	srv, _, _ := setupTestServer(t)

	logger := slog.Default()
	_, err := LoadAndRegister(srv, dir, logger)
	if err != nil {
		t.Fatalf("LoadAndRegister returned error: %v", err)
	}

	required := RequiredParams("sandbox")
	if len(required) != 0 {
		t.Errorf("expected empty or nil required params, got: %v", required)
	}
}

// TestEngineNamesReturnsSortedCopy verifies that EngineNames returns a sorted
// list and that modifying the returned slice does not affect future calls.
func TestEngineNamesReturnsSortedCopy(t *testing.T) {
	dir := t.TempDir()
	createTestEngineFile(t, dir, "charlie", `{"engine":"charlie","params":{}}`)
	createTestEngineFile(t, dir, "alpha", `{"engine":"alpha","params":{}}`)
	createTestEngineFile(t, dir, "bravo", `{"engine":"bravo","params":{}}`)

	srv, _, _ := setupTestServer(t)

	logger := slog.Default()
	_, err := LoadAndRegister(srv, dir, logger)
	if err != nil {
		t.Fatalf("LoadAndRegister returned error: %v", err)
	}

	names := EngineNames()
	expected := []string{"alpha", "bravo", "charlie"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d engines, got %d", len(expected), len(names))
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("expected names[%d]=%q, got %q", i, name, names[i])
		}
	}

	// Mutate the returned slice and verify EngineNames returns a fresh copy
	names[0] = "MUTATED"
	freshNames := EngineNames()
	if freshNames[0] == "MUTATED" {
		t.Error("EngineNames() returned the same slice reference — should return a copy")
	}
	if freshNames[0] != "alpha" {
		t.Errorf("expected freshNames[0]=\"alpha\", got %q", freshNames[0])
	}
}
