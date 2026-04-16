package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewCORSConfigDefaultOrigin(t *testing.T) {
	cfg := NewCORSConfig("*")
	if len(cfg.AllowedOrigins) != 1 {
		t.Fatalf("expected 1 origin, got %d", len(cfg.AllowedOrigins))
	}
	if cfg.AllowedOrigins[0] != "*" {
		t.Fatalf("expected origin *, got %s", cfg.AllowedOrigins[0])
	}
}

func TestNewCORSConfigEmptyOrigin(t *testing.T) {
	cfg := NewCORSConfig("")
	if len(cfg.AllowedOrigins) != 1 {
		t.Fatalf("expected 1 origin for empty string, got %d", len(cfg.AllowedOrigins))
	}
	if cfg.AllowedOrigins[0] != "*" {
		t.Fatalf("expected origin * for empty string, got %s", cfg.AllowedOrigins[0])
	}
}

func TestNewCORSConfigCustomOrigins(t *testing.T) {
	cfg := NewCORSConfig("https://example.com, https://other.com")
	if len(cfg.AllowedOrigins) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(cfg.AllowedOrigins))
	}
	if cfg.AllowedOrigins[0] != "https://example.com" {
		t.Fatalf("expected first origin https://example.com, got %s", cfg.AllowedOrigins[0])
	}
	if cfg.AllowedOrigins[1] != "https://other.com" {
		t.Fatalf("expected second origin https://other.com, got %s", cfg.AllowedOrigins[1])
	}
}

func TestCORSMiddlewareDefaultOrigin(t *testing.T) {
	cfg := NewCORSConfig("*")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := corsMiddleware(cfg, inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin: *, got %s", origin)
	}
}

func TestCORSMiddlewareCustomOrigin(t *testing.T) {
	cfg := NewCORSConfig("https://example.com")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := corsMiddleware(cfg, inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "https://example.com" {
		t.Fatalf("expected Access-Control-Allow-Origin: https://example.com, got %s", origin)
	}
}

func TestCORSMiddlewareHeaders(t *testing.T) {
	cfg := NewCORSConfig("*")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := corsMiddleware(cfg, inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Fatalf("expected Access-Control-Allow-Methods to include standard methods, got %s", methods)
	}

	headers := w.Header().Get("Access-Control-Allow-Headers")
	if headers != "Content-Type, Authorization, Mcp-Session-Id" {
		t.Fatalf("expected Access-Control-Allow-Headers with Authorization, got %s", headers)
	}

	credentials := w.Header().Get("Access-Control-Allow-Credentials")
	if credentials != "true" {
		t.Fatalf("expected Access-Control-Allow-Credentials: true, got %s", credentials)
	}
}

func TestCORSMiddlewarePreflight(t *testing.T) {
	cfg := NewCORSConfig("*")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called for OPTIONS")
	})
	handler := corsMiddleware(cfg, inner)

	req := httptest.NewRequest(http.MethodOptions, "/mcp", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for OPTIONS, got %d", w.Code)
	}

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin: *, got %s", origin)
	}

	methods := w.Header().Get("Access-Control-Allow-Methods")
	if methods == "" {
		t.Fatal("expected Access-Control-Allow-Methods header on preflight")
	}
}
