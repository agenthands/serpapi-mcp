// Package middleware provides HTTP middleware for request correlation and tracing.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// CorrelationIDHeader is the HTTP header name for correlation IDs.
// Clients may set this header to provide their own correlation ID;
// if absent, the middleware generates a new one.
const CorrelationIDHeader = "X-Correlation-ID"

// correlationIDKey is the context key used to store the correlation ID.
type correlationIDKey string

const correlationKey correlationIDKey = "correlation_id"

// CorrelationIDMiddleware wraps the next handler to inject a correlation ID
// into each request's context and response headers. If the request already
// contains an X-Correlation-ID header (client-provided), that value is used;
// otherwise a new 32-char hex ID is generated using crypto/rand.
func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(CorrelationIDHeader)
		if id == "" {
			id = generateCorrelationID()
		}

		ctx := context.WithValue(r.Context(), correlationKey, id)
		w.Header().Set(CorrelationIDHeader, id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CorrelationIDFromContext extracts the correlation ID from the given context.
// Returns an empty string if no ID is present.
func CorrelationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(correlationKey).(string)
	return id
}

// generateCorrelationID creates a new 32-character lowercase hex string
// using crypto/rand for secure, unique IDs.
func generateCorrelationID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
