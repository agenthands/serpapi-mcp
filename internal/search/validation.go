// Package search provides input validation for the SerpApi search tool.
package search

import (
	"fmt"
	"sort"
	"strings"

	"github.com/agenthands/serpapi-mcp/internal/engines"
)

// ValidateEngine checks that the given engine name is valid (exists in loaded engines).
// Returns an error listing available engines if the name is not found.
func ValidateEngine(engine string) error {
	validNames := engines.EngineNames()
	if validNames == nil {
		return fmt.Errorf("invalid_engine: engine '%s' not found. No engines loaded", engine)
	}

	// EngineNames returns a sorted list; use binary search
	idx := sort.SearchStrings(validNames, engine)
	if idx < len(validNames) && validNames[idx] == engine {
		return nil
	}

	return fmt.Errorf("invalid_engine: engine '%s' not found. Available engines: %s",
		engine, strings.Join(validNames, ", "))
}

// ValidateMode checks that the mode is either "complete" or "compact" (D-10).
// Returns an error with a clear message listing valid modes if invalid.
func ValidateMode(mode string) error {
	if mode == "complete" || mode == "compact" {
		return nil
	}
	return fmt.Errorf("invalid_mode: mode must be 'complete' or 'compact', got '%s'", mode)
}

// ValidateRequiredParams checks that all required parameters for the given engine
// are present in the params map. Returns an error listing missing param names.
// If the engine is not found (RequiredParams returns nil), validation is skipped
// since the engine name will be caught by ValidateEngine.
func ValidateRequiredParams(engine string, params map[string]any) error {
	required := engines.RequiredParams(engine)
	if required == nil {
		return nil
	}

	var missing []string
	for _, name := range required {
		if _, exists := params[name]; !exists {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing_params: missing required parameter(s) %s for engine '%s'",
			strings.Join(missing, ", "), engine)
	}

	return nil
}
