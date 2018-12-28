package metrics

/*
Usage Example:

1. Initialize a Metric instance as a global var:

var (
	MetricErrCounter = metrics.NewMetric(&metrics.Vector{
		Type: metrics.TypeCounter,
		Name: "service_error_counter",
		Desc: "service error counts",
		Labels: []string{"action", "error"},
	})
	...
)

2. Register global metric in function:

func main() {
	...

	vectors := metrics.NewMetricVectors([]*metrics.Metric{
		MetricErrCounter,
		...
	}, prependLabels...)

	metrics.Register(vectors...)
	metrics.Serve()

	...
}

3. (optional) Note from above, you could prepend a fixed set of labels to each metrics using this library.
	For example, the following code will prepend NSQ consumer setting to each metrics as their common labels.

	prependLabels := []metrics.PrependLabelMap{
		metrics.PrependLabelMap{
			Label: "input_topic", Value: options.Indexer.NSQInputTopic,
		}, metrics.PrependLabelMap{
			Label: "max_in_flight", Value: strconv.Itoa(options.Indexer.NSQMaxInFlight),
		}, metrics.PrependLabelMap{
			Label: "batch_size", Value: strconv.Itoa(options.Indexer.ConsumerBatchSize),
		}, metrics.PrependLabelMap{
			Label: "pool_size", Value: strconv.Itoa(options.Indexer.ConsumerPoolSize),
		}
	}

4. Use the following method to report metrics:

	metric.AddOne(labels...)
	metric.Add(val, labels...)
	metric.Set(val, labels...)
	metric.SinceStart(start, labels...)
	metric.SinceStartPerItem(start, batchSize, labels...)

*/

import (
	"strings"
	"time"
)

const milisecondSuffix = "_ms"

// PrependLabelFunc defines a function type that will prepend a static set of labels to given labels arg.
type PrependLabelFunc func(labels []string) []string

// Metric struct wraps Vector so that we can define our own reporter with shared labels.
type Metric struct {
	Vector               *Vector          `json:"-"`
	PrependedLabelValues PrependLabelFunc `json:"-"`
}

// NewMetric constructs a Metric instance.
func NewMetric(vector *Vector) *Metric {
	return &Metric{
		Vector: vector,
	}
}

// SetBuckets sets buckets for the metric vector
func (m *Metric) SetBuckets(buckets []float64) {
	m.Vector.Buckets = buckets
}

// AddOne report metrics counter by adding 1
func (m *Metric) AddOne(labels ...string) {
	if m == nil {
		return
	}
	if m.PrependedLabelValues == nil {
		Add(m.Vector.Name, float64(1), labels...)
		return
	}
	Add(m.Vector.Name, float64(1), m.PrependedLabelValues(labels)...)
}

// Add reports metrics counter by adding given val
func (m *Metric) Add(val float64, labels ...string) {
	if m == nil {
		return
	}
	if m.PrependedLabelValues == nil {
		Add(m.Vector.Name, val, labels...)
		return
	}
	Add(m.Vector.Name, val, m.PrependedLabelValues(labels)...)
}

// Set reports metrics gauge by setting given val
func (m *Metric) Set(val float64, labels ...string) {
	if m == nil {
		return
	}
	if m.PrependedLabelValues == nil {
		Set(m.Vector.Name, val, labels...)
		return
	}
	Set(m.Vector.Name, val, m.PrependedLabelValues(labels)...)
}

// SinceStart reports timer duration since given start time
func (m *Metric) SinceStart(start time.Time, labels ...string) {
	if m == nil {
		return
	}
	duration := float64(time.Since(start).Nanoseconds())
	if strings.HasSuffix(m.Vector.Name, milisecondSuffix) {
		duration = duration / float64(time.Millisecond)
	} else {
		duration = duration / float64(time.Second)
	}
	if m.PrependedLabelValues == nil {
		Add(m.Vector.Name, duration, labels...)
		return
	}
	Add(m.Vector.Name, duration, m.PrependedLabelValues(labels)...)
}

// SinceStartPerItem reports timer duration since given start time divided by given batchSize
func (m *Metric) SinceStartPerItem(start time.Time, batchSize int, labels ...string) {
	if m == nil {
		return
	}
	durationPerItem := float64(time.Since(start).Nanoseconds()) / float64(batchSize)
	if strings.HasSuffix(m.Vector.Name, milisecondSuffix) {
		durationPerItem = durationPerItem / float64(time.Millisecond)
	} else {
		durationPerItem = durationPerItem / float64(time.Second)
	}
	if m.PrependedLabelValues == nil {
		Add(m.Vector.Name, durationPerItem, labels...)
		return
	}
	Add(m.Vector.Name, durationPerItem, m.PrependedLabelValues(labels)...)
}

// PrependLabelMap defines a stuct of string lable and value to be prepended to the metric vector.
type PrependLabelMap struct {
	Label string
	Value string
}

// NewMetricVectors constructs a list of Vector based on given Metric list with prepended lables.
func NewMetricVectors(metricsList []*Metric, prependLabels ...PrependLabelMap) []*Vector {
	// get prepend label names and values
	var labels, values []string
	for _, labelMap := range prependLabels {
		labels = append(labels, labelMap.Label)
		values = append(values, labelMap.Value)
	}

	// prepend label
	vectors := []*Vector{}
	for _, metric := range metricsList {
		metric.Vector.Labels = append(labels, metric.Vector.Labels...)
		metric.PrependedLabelValues = func(labelValues []string) []string {
			return append(values, labelValues...)
		}
		vectors = append(vectors, metric.Vector)
	}

	return vectors
}
