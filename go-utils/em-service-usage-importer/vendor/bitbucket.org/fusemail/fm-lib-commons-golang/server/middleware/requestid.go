package middleware

import (
	"context"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

// RequestIDHeaderKey defines request id header key string
const RequestIDHeaderKey = "request_id"

// RequestIDContextKey defines request id context key instance
const RequestIDContextKey = ContextKey(RequestIDHeaderKey)

// RequestID is a middleware that addes x-request-id header to each request
type RequestID struct{}

// NewRequestID constructs RequestID instance
func NewRequestID() *RequestID {
	return &RequestID{}
}

func (reqid *RequestID) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	uid, err := uuid.NewV1()
	if err != nil {
		panic("failed to generate a uuid")
	}
	id := uid.String()
	// add request id to request context
	ctx := context.WithValue(r.Context(), RequestIDContextKey, id)
	r = r.WithContext(ctx)
	// store request id in http header
	r.Header.Set(RequestIDHeaderKey, id)
	rw.Header().Set(RequestIDHeaderKey, id)
	next(rw, r)
}

// GetRequestID returns request id from context.
// If not found then returns empty string.
func GetRequestID(ctx context.Context) string {
	ID := ctx.Value(RequestIDContextKey)
	if s, ok := ID.(string); ok {
		return s
	}
	return ""
}
