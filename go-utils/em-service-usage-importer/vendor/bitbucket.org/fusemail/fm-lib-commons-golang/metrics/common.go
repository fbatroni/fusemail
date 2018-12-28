package metrics

// This file defines some common default metrics shared by all services

// metrics vectors
var (
	ErrorCounter = NewMetric(&Vector{
		Type:   TypeCounter,
		Name:   "error_counter",
		Desc:   "Serivce error counter",
		Labels: []string{"action", "error"},
	})
	GRPCRequestCounter = NewMetric(&Vector{
		Type:   TypeCounter,
		Name:   "grpc_request_total",
		Desc:   "total GRPC service requests",
		Labels: []string{"addr", "action"},
	})
	GRPCRequestTimer = NewMetric(&Vector{
		Type:    TypeHistogram,
		Name:    "grpc_request_duration_ms",
		Desc:    "GRPC service response time in miliseconds",
		Labels:  []string{"addr", "action"},
		Buckets: []float64{0.1, 0.25, 0.3, 0.45, 0.5, 1, 2.5, 5, 7.5, 10, 15, 20, 25, 50, 75, 100, 150, 200, 250, 500, 750, 1000, 2500, 5000},
	})
	EntryProcessDurationTimer = NewMetric(&Vector{
		Type:    TypeHistogram,
		Name:    "nsq_entry_process_duration_ms",
		Desc:    "nsq entry process duration in milliseconds",
		Labels:  []string{"input_topic", "batch_size", "pool_size", "result"},
		Buckets: []float64{1, 2.5, 5, 7.5, 10, 25, 50, 75, 100, 150, 200, 250, 500, 1000, 2500, 5000, 7500, 10000, 30000, 60000, 120000, 180000, 300000, 450000, 600000, 750000, 900000},
	})
	EntryProcessCounter = NewMetric(&Vector{
		Type:   TypeCounter,
		Name:   "nsq_entry_process_counter",
		Desc:   "nsq entry process counter",
		Labels: []string{"input_topic", "batch_size", "pool_size", "result"},
	})
	EntryFilterCounter = NewMetric(&Vector{
		Type:   TypeCounter,
		Name:   "nsq_entry_filter_counter",
		Desc:   "nsq entry filter counter",
		Labels: []string{"input_topic", "filter_type"},
	})
)

// ToMetricLabel shortens given string to specified length if longer.
// this is often used to cut short error messages whose first 25 characters are usually a fixed set.
func ToMetricLabel(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length]
}
