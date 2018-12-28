/*
Package middleware provides common reusable middleware:
	recovery, requestid, logger, metrics, limiter.
Instruments metrics:
	http_request_total, http_request_bytes, http_request_duration_seconds.
*/
package middleware

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

// Package logger, set with SetLogger.
var log = logrus.StandardLogger()

// SetLogger overrides the package logger.
func SetLogger(logger *logrus.Logger) {
	log = logger
}

// ContextKey is just a named string type to define context key value.
type ContextKey string

// Common provides the default common middleware by negroni.
// Negroni is a lightweight middleware, bring your own router.
func Common() *negroni.Negroni {
	return negroni.New(
		NewRecovery(),

		// Before NewLogger.
		NewRequestID(),
		NewTransactionID(),

		NewLogger(),
		NewMetrics(),
	)
}

// GetContextLogger returns the context logger with all pertinent keys and values.
func GetContextLogger(ctx context.Context) *logrus.Entry {
	logger := logrus.NewEntry(log) // Empty by default.

	requestID := GetRequestID(ctx)
	transactionID := GetTransactionID(ctx)
	if transactionID != "" {
		if requestID == "" {
			requestID = transactionID
		} else {
			requestID = fmt.Sprintf("%s-%s", transactionID, requestID)
		}
	}

	for _, each := range []struct {
		key, val string
	}{
		{key: RequestIDHeaderKey, val: requestID},
		{key: TransactionIDHeaderKey, val: transactionID},
	} {
		if each.val != "" {
			logger = logger.WithField(each.key, each.val)
		}
	}

	// logger.Info("GetContextLogger")
	return logger
}
