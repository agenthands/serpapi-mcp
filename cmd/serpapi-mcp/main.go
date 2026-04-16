package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/agenthands/serpapi-mcp/internal/engines"
	"github.com/agenthands/serpapi-mcp/internal/server"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// CLI flags with defaults matching Python server
	hostFlag := flag.String("host", envOr("MCP_HOST", "0.0.0.0"), "Host to bind the server to")
	portFlag := flag.Int("port", envIntOr("MCP_PORT", 8000), "Port to bind the server to")
	corsOriginsFlag := flag.String("cors-origins", envOr("MCP_CORS_ORIGINS", "*"), "Comma-separated list of allowed CORS origins")
	authDisabledFlag := flag.Bool("auth-disabled", envBoolOr("MCP_AUTH_DISABLED", false), "Disable API key authentication (for testing)")
	enginesDirFlag := flag.String("engines-dir", envOr("ENGINES_DIR", "engines"), "Path to directory containing engine JSON schemas")
	flag.Parse()

	// Print startup banner
	fmt.Printf("serpapi-mcp %s (commit: %s, built: %s)\n", version, commit, date)

	cfg := server.Config{
		Host:         *hostFlag,
		Port:         *portFlag,
		CorsOrigins:  *corsOriginsFlag,
		AuthDisabled: *authDisabledFlag,
	}

	mcpServer := server.NewMCPServer(cfg, version)

	// Load engine schemas and register MCP resources (fail-fast on corrupt/missing JSON)
	engineCount, err := engines.LoadAndRegister(mcpServer.MCPServer, *enginesDirFlag, slog.Default())
	if err != nil {
		slog.Error("failed to load engine schemas", "error", err)
		os.Exit(1)
	}
	mcpServer.SetEngineCount(engineCount)

	// Set up signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := mcpServer.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "server failed: %v\n", err)
		os.Exit(1)
	}
}

// envOr returns the value of the environment variable named key, or the
// provided fallback if the variable is not set or empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envIntOr returns the integer value of the environment variable named key,
// or the provided fallback if the variable is not set, empty, or not a valid integer.
func envIntOr(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// envBoolOr returns the boolean value of the environment variable named key,
// or the provided fallback if the variable is not set. Accepts "1", "true", "yes"
// as truthy values; everything else is falsy.
func envBoolOr(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return strings.EqualFold(v, "1") || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
}
