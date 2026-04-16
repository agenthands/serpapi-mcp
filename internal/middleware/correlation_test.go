package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCorrelationIDMiddlewareGeneratesID verifies that the middleware
// generates a correlation ID for each request and stores it in context.
func TestCorrelationIDMiddlewareGeneratesID(t *testing.T) {
	var capturedID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = CorrelationIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := CorrelationIDMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedID == "" {
		t.Fatal("expected correlation ID to be generated, got empty string")
	}

	// Should be a 32-char hex string (16 bytes hex-encoded)
	if len(capturedID) != 32 {
		t.Errorf("expected 32-char hex ID, got %d chars: %s", len(capturedID), capturedID)
	}

	// Should only contain hex characters
	for _, c := range capturedID {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("expected hex-only ID, got char %c in %s", c, capturedID)
			break
		}
	}
}

// TestCorrelationIDMiddlewareSetsResponseHeader verifies that the middleware
// sets the X-Correlation-ID response header.
func TestCorrelationIDMiddlewareSetsResponseHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CorrelationIDMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	respHeader := rec.Header().Get(CorrelationIDHeader)
	if respHeader == "" {
		t.Fatal("expected X-Correlation-ID response header to be set")
	}

	if len(respHeader) != 32 {
		t.Errorf("expected 32-char hex ID in header, got %d chars: %s", len(respHeader), respHeader)
	}
}

// TestCorrelationIDMiddlewareUsesClientID verifies that when a client
// provides an X-Correlation-ID header, the middleware uses that value.
func TestCorrelationIDMiddlewareUsesClientID(t *testing.T) {
	clientID := "client-provided-trace-id-12345"

	var capturedID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = CorrelationIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := CorrelationIDMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(CorrelationIDHeader, clientID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedID != clientID {
		t.Errorf("expected client-provided ID %q, got %q", clientID, capturedID)
	}

	respHeader := rec.Header().Get(CorrelationIDHeader)
	if respHeader != clientID {
		t.Errorf("expected response header %q, got %q", clientID, respHeader)
	}
}

// TestCorrelationIDFromContextMissing verifies that CorrelationIDFromContext
// returns empty string when no ID is present in context.
func TestCorrelationIDFromContextMissing(t *testing.T) {
	id := CorrelationIDFromContext(nil)
	if id != "" {
		t.Errorf("expected empty string for nil context, got %q", id)
	}
}

// TestCorrelationIDUniquePerRequest verifies that each request gets
// a unique correlation ID (IDs differ between concurrent requests).
func TestCorrelationIDUniquePerRequest(t *testing.T) {
	var ids []string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := CorrelationIDFromContext(r.Context())
		ids = append(ids, id)
		w.WriteHeader(http.StatusOK)
	})

	handler := CorrelationIDMiddleware(next)

	// Make 10 requests
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// Verify all IDs are unique
	seen := make(map[string]bool)
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate correlation ID found: %s", id)
		}
		seen[id] = true
	}

	if len(seen) != 10 {
		t.Errorf("expected 10 unique IDs, got %d", len(seen))
	}
}

// TestCorrelationIDContainsOnlyHex verifies that generated IDs
// only contain lowercase hex characters.
func TestCorrelationIDContainsOnlyHex(t *testing.T) {
	var capturedID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = CorrelationIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := CorrelationIDMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !isHex(capturedID) {
		t.Errorf("expected hex-only ID, got: %s", capturedID)
	}
}

// isHex checks if a string contains only lowercase hex characters.
func isHex(s string) bool {
	for _, c := range s {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}
