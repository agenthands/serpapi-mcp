package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/agenthands/serpapi-mcp/internal/engines"
	"github.com/agenthands/serpapi-mcp/internal/search"
	"github.com/agenthands/serpapi-mcp/internal/server"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "server failed: %v\n", err)
		os.Exit(1)
	}
}

// run contains the server startup logic extracted from main() for testability.
// ctx controls graceful shutdown, args are CLI arguments (without program name).
func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("serpapi-mcp", flag.ContinueOnError)
	hostFlag := fs.String("host", envOr("MCP_HOST", "0.0.0.0"), "Host to bind the server to")
	portFlag := fs.Int("port", envIntOr("MCP_PORT", 8000), "Port to bind the server to")
	corsOriginsFlag := fs.String("cors-origins", envOr("MCP_CORS_ORIGINS", "*"), "Comma-separated list of allowed CORS origins")
	authDisabledFlag := fs.Bool("auth-disabled", envBoolOr("MCP_AUTH_DISABLED", false), "Disable API key authentication (for testing)")
	enginesDirFlag := fs.String("engines-dir", envOr("ENGINES_DIR", "engines"), "Path to directory containing engine JSON schemas")

	if err := fs.Parse(args); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "serpapi-mcp %s (commit: %s, built: %s)\n", version, commit, date)

	cfg := server.Config{
		Host:         *hostFlag,
		Port:         *portFlag,
		CorsOrigins:  *corsOriginsFlag,
		AuthDisabled: *authDisabledFlag,
	}

	mcpServer := server.NewMCPServer(cfg, version)

	engineCount, err := engines.LoadAndRegister(mcpServer.MCPServer, *enginesDirFlag, slog.Default())
	if err != nil {
		return fmt.Errorf("failed to load engine schemas: %w", err)
	}
	mcpServer.SetEngineCount(engineCount)

	search.RegisterSearchTool(mcpServer.MCPServer, slog.Default())
	slog.Info("search tool registered")

	return mcpServer.Run(ctx)
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
