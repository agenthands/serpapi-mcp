// Package server provides the MCP server implementation.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Config holds the server configuration.
type Config struct {
	Host         string
	Port         int
	CorsOrigins  string
	AuthDisabled bool
}

// MCPServer wraps the MCP server with HTTP transport, healthcheck, and CORS.
type MCPServer struct {
	MCPServer   *mcp.Server
	HTTPHandler *mcp.StreamableHTTPHandler
	Config      Config
	httpServer  *http.Server
	logger      *slog.Logger
	version     string
	engineCount int
}

// healthResponse is the JSON response body for the /health endpoint.
type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// NewMCPServer creates a new MCP server with Streamable HTTP transport.
func NewMCPServer(cfg Config, version string) *MCPServer {
	logger := slog.Default()

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "SerpApi MCP Server",
			Version: version,
		},
		&mcp.ServerOptions{
			Logger: logger,
		},
	)

	httpHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return mcpServer
		},
		&mcp.StreamableHTTPOptions{
			Stateless:                  true,
			JSONResponse:               true,
			Logger:                     logger,
			DisableLocalhostProtection: true,
		},
	)

	return &MCPServer{
		MCPServer:   mcpServer,
		HTTPHandler: httpHandler,
		Config:      cfg,
		logger:      logger,
		version:     version,
	}
}

// buildHandler constructs the HTTP handler chain: CORS → mux (with /mcp and /health routes).
func (s *MCPServer) buildHandler() http.Handler {
	mux := http.NewServeMux()

	// MCP Streamable HTTP endpoint
	mux.Handle("/mcp", s.HTTPHandler)

	// Healthcheck endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := healthResponse{
			Status:  "healthy",
			Service: "SerpApi MCP Server",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode health response", "error", err)
		}
	})

	// Wrap with auth middleware (skipped if AuthDisabled)
	authenticated := authOrPassthrough(s.Config.AuthDisabled, mux)

	// Wrap with CORS middleware
	corsCfg := NewCORSConfig(s.Config.CorsOrigins)
	return corsMiddleware(corsCfg, authenticated)
}

// SetEngineCount sets the number of loaded engines for startup logging.
func (s *MCPServer) SetEngineCount(count int) {
	s.engineCount = count
}

// Run starts the MCP server and blocks until the context is cancelled.
// It listens for SIGINT/SIGTERM via the provided context and shuts down gracefully.
func (s *MCPServer) Run(ctx context.Context) error {
	handler := s.buildHandler()

	addr := fmt.Sprintf("%s:%d", s.Config.Host, s.Config.Port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Start listening
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Log actual port (useful when Port=0 for random port)
	actualAddr := listener.Addr().String()
	s.logger.Info("SerpApi MCP Server starting",
		"address", actualAddr,
		"version", s.version,
		"engines_loaded", s.engineCount,
	)

	// Start serving in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
		close(serverErr)
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("shutdown signal received, shutting down gracefully...")
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	s.logger.Info("server shutdown complete")
	return nil
}
