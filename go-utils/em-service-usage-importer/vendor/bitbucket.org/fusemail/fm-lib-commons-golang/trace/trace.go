/*
Package trace provides generic tracing capabilities.

Embed Tracers in target struct to be traced (so it inherits AddTracers and Trace funcs, along with "tracers" field):
	type Any struct {
		trace.Tracers
	}

Then trace in any function:
	t := trace.New(actionAny)
	defer any.Trace(t)

And caller can add tracers as required:
	job.AddTracers(trace.Logger(), ..)
*/
package trace

import (
	"time"

	"bitbucket.org/fusemail/fm-lib-commons-golang/metrics"
	log "github.com/sirupsen/logrus"
)

// Tracers represents struct being traced, must embed.
type Tracers struct {
	tracers []Func
}

// AddTracers appends trace functions to the traced struct.
func (t *Tracers) AddTracers(tracers ...Func) {
	t.tracers = append(t.tracers, tracers...)
}

// GetTracers returns trace functions.
func (t *Tracers) GetTracers() []Func {
	return t.tracers
}

// Trace applies tracing to the traced struct.
func (t *Tracers) Trace(trace *Trace) {
	if t == nil {
		return
	}
	for _, tracer := range t.tracers {
		tracer(trace)
	}
}

// Trace traces according to functions.
type Trace struct {
	Action string
	Fields map[string]interface{}
	Start  time.Time
	Error  error

	Logger      *log.Entry
	LogDebug    bool
	LogDuration bool
}

// New constructs new trace instance.
func New(action string) *Trace {
	return &Trace{
		Action: action,
		Start:  time.Now(),
		Fields: make(map[string]interface{}),
	}
}

// Func represents trace function.
type Func func(*Trace)

// Logger provides trace function to log.
func Logger() Func {
	return func(t *Trace) {
		if t.Logger == nil {
			t.Logger = log.NewEntry(log.StandardLogger())
		}
		logField := func(key string, val interface{}) {
			t.Logger = t.Logger.WithField(key, val)
		}

		for key, val := range t.Fields {
			logField(key, val)
		}

		if t.LogDuration && !t.Start.IsZero() {
			logField("start_time", t.Start.Format(time.RFC3339Nano))

			// Duration in rounded milliseconds.
			ms := time.Since(t.Start).Nanoseconds() / 1e6
			logField("duration_ms", ms)
		}

		if t.Error != nil {
			logField("error", t.Error.Error())
			t.Logger.Error(t.Action)
			return
		}

		if t.LogDebug {
			t.Logger.Debug(t.Action)
		} else {
			t.Logger.Info(t.Action)
		}
	}
}

// ErrorMetric reports to error metric.
func ErrorMetric(counter *metrics.Metric) Func {
	return func(t *Trace) {
		if t.Error != nil {
			errStr := metrics.ToMetricLabel(t.Error.Error(), 25)
			counter.AddOne(t.Action, errStr)
		}
	}
}
