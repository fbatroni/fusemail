package metricvec

import "bitbucket.org/fusemail/fm-lib-commons-golang/metrics"

// metrics vectors
var (
	MetricSQLQueryCounter = metrics.NewMetric(&metrics.Vector{
		Type:   metrics.TypeCounter,
		Name:   "sql_query_counter",
		Desc:   "DB query counter",
		Labels: []string{"query"},
	})
	MetricSQLQueryTimer = metrics.NewMetric(&metrics.Vector{
		Type:   metrics.TypeHistogram,
		Name:   "sql_query_duration_ms",
		Desc:   "DB query duration in milliseconds",
		Labels: []string{"query"},
	})
	MetricActionDurationTimer = metrics.NewMetric(&metrics.Vector{
		Type:   metrics.TypeHistogram,
		Name:   "action_duration_sec",
		Desc:   "verifier ation duration in seconds",
		Labels: []string{"action"},
	})
	MetricDownloadTimer = metrics.NewMetric(&metrics.Vector{
		Type:   metrics.TypeHistogram,
		Name:   "download_duration_sec",
		Desc:   "verifier download duration in seconds",
		Labels: []string{"download", "source"},
	})
)
