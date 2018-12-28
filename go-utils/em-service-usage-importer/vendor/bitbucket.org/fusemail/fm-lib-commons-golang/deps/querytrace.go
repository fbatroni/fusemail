package deps

import (
	"time"

	"bitbucket.org/fusemail/fm-lib-commons-golang/metrics"
)

// QueryTrace contains action details to apply trace function to.
type QueryTrace struct {
	Action string
	Start  time.Time
	Error  error
}

// QueryTraceFunc is a closure function that implements on a QueryTracer instance. e.g. reportMetric, logging etc.
type QueryTraceFunc func(*QueryTrace)

// ReportQueryMetrics reports both counter and timer metrics.
func ReportQueryMetrics(counter, timer *metrics.Metric) QueryTraceFunc {
	return func(t *QueryTrace) {
		counter.AddOne(t.Action)
		timer.SinceStart(t.Start, t.Action)
	}
}
