package middleware

import (
	"context"
	"net/http"
)

// TransactionIDHeaderKey defines transaction id header key.
const TransactionIDHeaderKey = "X-Fm-Transaction-Id"

// TransactionIDContextKey defines transaction id context key instance.
const TransactionIDContextKey = ContextKey(TransactionIDHeaderKey)

// TransactionID is middleware that adds transaction id header to each request
type TransactionID struct{}

// NewTransactionID constructs TransactionID instance.
func NewTransactionID() *TransactionID {
	return &TransactionID{}
}

// ServeHTTP handles web request as middleware for transaction id.
func (t *TransactionID) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// If found in request header, then add to context and assign to response header.
	if val := r.Header.Get(TransactionIDHeaderKey); val != "" {
		ctx := context.WithValue(r.Context(), TransactionIDContextKey, val)
		r = r.WithContext(ctx)
		rw.Header().Set(TransactionIDHeaderKey, val)
	}
	next(rw, r)
}

// GetTransactionID returns transaction id from context, or empty string if not found.
func GetTransactionID(ctx context.Context) string {
	val := ctx.Value(TransactionIDContextKey)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}
