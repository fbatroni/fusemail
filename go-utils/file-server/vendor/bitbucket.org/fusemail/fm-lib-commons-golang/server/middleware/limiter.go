package middleware

import (
	"net/http"
	"time"
)

// DefaultHTTPTimeoutStatus contains timeout HTTP status code
const DefaultHTTPTimeoutStatus = http.StatusTooManyRequests

// Limiter is a middleware handler that limits the request as it goes in.
type Limiter struct {
	Limit             int
	Chan              chan struct{}
	Timeout           time.Duration
	TimeoutHTTPStatus int
}

// NewLimiter constructs limiter middleware instances that take in the maximum connection value.
func NewLimiter(maxConnections int) *Limiter {
	l := &Limiter{
		Limit:             maxConnections,
		TimeoutHTTPStatus: DefaultHTTPTimeoutStatus,
		Timeout:           5 * time.Second,
	}
	l.Chan = make(chan struct{}, maxConnections)
	return l
}

func (l *Limiter) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// By default, limiter is disabled.
	if l.Limit == 0 {
		next(rw, r)
		return
	}

	select {
	case <-time.After(l.Timeout):
		http.Error(rw, "timed out waiting for a thread, hit max connections", l.TimeoutHTTPStatus)
		return
	case l.Chan <- struct{}{}:
		next(rw, r)
		<-l.Chan
	}
}
