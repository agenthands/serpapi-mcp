// Package engines provides engine schema loading and resource serving.
package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// engineNamePattern validates engine filenames: only [a-z0-9_]+.json allowed.
var engineNamePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

// engineSchema holds a parsed engine JSON schema.
type engineSchema struct {
	Engine string          `json:"engine"`
	Raw    json.RawMessage `json:"-"`
}

// LoadAndRegister loads all engine JSON schemas from the given directory,
// validates them, and registers MCP resources for engine discovery.
// It returns the number of engines loaded or an error if validation fails.
func LoadAndRegister(srv *mcp.Server, enginesDir string, logger *slog.Logger) (int, error) {
	entries, err := os.ReadDir(enginesDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read engines directory %q: %w", enginesDir, err)
	}

	schemas := make(map[string]*engineSchema)
	var engineNames []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if !strings.HasSuffix(fileName, ".json") {
			continue
		}

		stem := strings.TrimSuffix(fileName, ".json")

		// Validate engine name: only [a-z0-9_]+ allowed
		if !engineNamePattern.MatchString(stem) {
			logger.Warn("skipping invalid engine filename", "file", fileName)
			continue
		}

		filePath := filepath.Join(enginesDir, fileName)
		data, err := os.ReadFile(filePath)
		if err != nil {
			return 0, fmt.Errorf("failed to read engine file %q: %w", filePath, err)
		}

		var schema engineSchema
		if err := json.Unmarshal(data, &schema); err != nil {
			return 0, fmt.Errorf("failed to parse engine JSON %q: %w", filePath, err)
		}
		schema.Raw = json.RawMessage(data)

		if schema.Engine != stem {
			return 0, fmt.Errorf("engine JSON %q has engine field %q, expected %q", fileName, schema.Engine, stem)
		}

		schemas[stem] = &schema
		engineNames = append(engineNames, stem)
	}

	sort.Strings(engineNames)

	// Register serpapi://engines index resource
	registerEnginesIndex(srv, engineNames, logger)

	// Register per-engine resources
	for _, name := range engineNames {
		registerEngineResource(srv, name, schemas[name], logger)
	}

	logger.Info("engines loaded", "count", len(engineNames))
	return len(engineNames), nil
}

// registerEnginesIndex registers the serpapi://engines resource that lists all engines.
func registerEnginesIndex(srv *mcp.Server, engineNames []string, logger *slog.Logger) {
	// Capture engineNames for the closure
	names := make([]string, len(engineNames))
	copy(names, engineNames)

	resourceURIs := make([]string, len(names))
	for i, name := range names {
		resourceURIs[i] = fmt.Sprintf("serpapi://engines/%s", name)
	}

	index := map[string]any{
		"count":     len(names),
		"engines":   names,
		"resources": resourceURIs,
		"schema": map[string]any{
			"note":              "Each engine resource uses a flat schema: params are engine-specific; common_params are shared SerpApi parameters.",
			"params_key":        "params",
			"common_params_key": "common_params",
		},
	}

	indexJSON, err := json.Marshal(index)
	if err != nil {
		logger.Error("failed to marshal engines index", "error", err)
		indexJSON = []byte(`{"error":"failed to marshal index"}`)
	}

	capturedJSON := string(indexJSON)

	srv.AddResource(
		&mcp.Resource{
			URI:         "serpapi://engines",
			Name:        "serpapi-engines-index",
			Description: "Index of available SerpApi engines and their resource URIs.",
			MIMEType:    "application/json",
		},
		func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{
						URI:      "serpapi://engines",
						MIMEType: "application/json",
						Text:     capturedJSON,
					},
				},
			}, nil
		},
	)
}

// registerEngineResource registers a per-engine MCP resource at serpapi://engines/{name}.
func registerEngineResource(srv *mcp.Server, name string, schema *engineSchema, logger *slog.Logger) {
	capturedText := string(schema.Raw)

	srv.AddResource(
		&mcp.Resource{
			URI:         fmt.Sprintf("serpapi://engines/%s", name),
			Name:        fmt.Sprintf("serpapi-engine-%s", name),
			Description: fmt.Sprintf("SerpApi engine specification for %s.", name),
			MIMEType:    "application/json",
		},
		func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{
						URI:      fmt.Sprintf("serpapi://engines/%s", name),
						MIMEType: "application/json",
						Text:     capturedText,
					},
				},
			}, nil
		},
	)
}
