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
		Buckets: []float64{0.0001, 0.00025, 0.0005, 0.00075, 0.001, 0.0025, 0.005, 0.0075, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.3, 0.45, 0.5, 1, 2.5, 5, 7.5, 10, 15, 20, 25, 50, 75, 100, 150, 200, 250, 500, 750, 1000},
	})
	EntryProcessDurationTimer = NewMetric(&Vector{
		Type:    TypeHistogram,
		Name:    "nsq_entry_process_duration_ms",
		Desc:    "nsq entry process duration in milliseconds",
		Labels:  []string{"input_topic", "batch_size", "pool_size", "result"},
		Buckets: []float64{0.001, 0.005, 0.01, 0.02, 0.05, 0.07, 0.1, 0.15, 0.2, 0.5, 0.75, 1, 1.5, 2, 2.5, 5, 7.5, 10},
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
