package handlers

import (
	"container/ring"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var logRing *LogRingHook

// LogRingHook is a log hook for log ring.
type LogRingHook struct {
	ring *ring.Ring
}

// newLogRingHook constructs log rings with the right size.
func newLogRingHook(numLines int) *LogRingHook {
	l := &LogRingHook{}
	l.ring = ring.New(numLines)
	return l
}

// Fire executes the hook.
func (hook *LogRingHook) Fire(entry *log.Entry) error {
	hook.ring.Value = entry
	hook.ring = hook.ring.Next()
	return nil
}

// Levels returns permitted log levels.
func (hook *LogRingHook) Levels() []log.Level {
	return log.AllLevels
}

// Log takes a number of log lines to keep in the ring buffer.
func Log(numLines int) http.Handler {
	logRing = newLogRingHook(numLines)
	log.AddHook(logRing)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ts = "2006-01-02T15:04:05.999999Z07:00"
		var formatter log.Formatter = &log.TextFormatter{DisableColors: true, FullTimestamp: true, TimestampFormat: ts}

		format := r.URL.Query().Get("format")

		if format == "json" {
			formatter = &log.JSONFormatter{TimestampFormat: ts}
		}

		logRing.ring.Do(func(x interface{}) {
			if x != nil {
				line, err := formatter.Format(x.(*log.Entry))

				if err != nil {
					return
				}

				w.Write([]byte(line))
			}
		})
	})
}
