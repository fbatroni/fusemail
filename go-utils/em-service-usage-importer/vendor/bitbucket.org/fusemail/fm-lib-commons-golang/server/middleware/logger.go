package middleware

import (
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

// Logger is a middleware handler that logs the request as it goes in and the response as it goes out.
type Logger struct{}

// NewLogger constructs Logger instance.
func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	proxy, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		proxy = r.RemoteAddr
	}

	ip := proxy

	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ip = forwardedFor
	}

	contextLogger := GetContextLogger(r.Context()).WithFields(logrus.Fields{
		"start":       start.Format(time.RFC3339),
		"client_ip":   ip,
		"proxy":       proxy,
		"server_name": r.Host,
		"method":      r.Method,
		"user_agent":  r.UserAgent(),
		"uri":         r.URL.Path,
		"uri_query":   r.URL.RawQuery,
	})

	contextLogger.Info("request started")

	next(rw, r)

	res := rw.(negroni.ResponseWriter)

	contextLogger.WithFields(logrus.Fields{
		"status":           res.Status(),
		"duration_seconds": time.Since(start).Seconds(),
	}).Info("request completed")
}
