/*
Package metrics is a concrete layer/wrapper for metrics, with prometheus as default Adapter.
Duties:
- Exposes Type* iota constants (VectorType types) to use as Vector's Type.
- Exposes Adaptor interface to plug any concrete Adapter implementation.
- Provides Prometheus Adapter, and uses it as default.
- Provides Mock Adapter, which satisfies Adaptor interface, but does nothing.
- Exposes polymorphic "Add" layer, which translates into concrete actions per Adapter.
- Avoids the need to import extra packages in your app, e.g. prometheus concrete implementation.
- Hides complexities of concrete Adapter implementations.
- Registers default metrics for other packages:
	server, client, health.
- Provides WebHandler to expose web endpoint.
- Sets up everything via Serve.
- Logs pertinent info.
*/
package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"bitbucket.org/fusemail/fm-lib-commons-golang/sys"
	"github.com/sirupsen/logrus"
)

// Types for valid Vector types.
const (
	TypeCounter VectorType = iota
	TypeGauge
	TypeSummary
	TypeHistogram
)

var (
	// Served indicates whether metrics has been served.
	Served bool

	// Stats contains metrics statistics.
	Stats = struct {
		Total int `json:"total"`
	}{}

	// Adapter holds the concrete adapter that satisfies Adaptor interface.
	// Use Prometheus as default Adapter.
	Adapter Adaptor = &PrometheusAdapter{}

	// Vectors holds classes/definitions by Name key.
	Vectors = make(map[string]*Vector)

	// Package logger, set with SetLogger.
	log = logrus.StandardLogger()

	// DefaultSizeBuckets is the default buckets for size/byte histogram
	DefaultSizeBuckets = []float64{512, 2048, 8192, 32768, 131072, 524288, 2097152, 8388608, 33554432, 134217728} // 0.5K/2K/8K/32K/128K/512K/2M/8M/32M/128M

	// DefaultTimeBuckets is the default buckets for time/seconds histogram
	DefaultTimeBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 120, 180, 300}
)

// SetLogger overrides the package logger.
func SetLogger(logger *logrus.Logger) {
	log = logger
}

// VectorType refers to valid Vector types.
type VectorType int

// Must use value receiver (instead of pointer) in order to handle the below switch,
// otherwise errors: mismatched types.
func (t VectorType) String() string {
	switch t {
	case TypeCounter:
		return "Counter"
	case TypeGauge:
		return "Gauge"
	case TypeSummary:
		return "Summary"
	case TypeHistogram:
		return "Histogram"
	default:
		return "INVALID"
	}
}

// MarshalJSON supports json log format.
func (t *VectorType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// Adaptor is the interface for concrete implementations.
type Adaptor interface {
	fmt.Stringer
	Serve()
	WebHandler() http.Handler
	Do(bool, *Vector, float64, ...string)
}

// Vector is the metric class/definition, with meta Labels.
// TBD - support for Objectives @ Summary, and Buckets @ Histogram?.
type Vector struct {
	Type    VectorType
	Name    string
	Desc    string
	Labels  []string    // Vectors (optional / nil).
	Buckets []float64   // Custom-defined buckets (optional, nil to use default)
	Ref     interface{} // Underlying reference instance as per Adapter.
}

func (v *Vector) String() string {
	return fmt.Sprintf(
		"{%T: %v (%v) %v}",
		v, v.Name, v.Type, v.Labels,
	)
}

// histogramBuckets holds custom buckets for default histogram metrics, i.e.
// web client and server, registered in metrics.Serve().
type histogramBuckets struct {
	BucketBytesServer   []float64
	BucketBytesClient   []float64
	BucketSecondsServer []float64
	BucketSecondsClient []float64
}

// newHistogramBuckets loads custom buckets from ENV to a histogramBuckets instance.
// If corresponding environmental variable(s) does not exist, use default buckets.
func newHistogramBuckets() *histogramBuckets {
	ret := &histogramBuckets{}

	strBuckets := os.Getenv("SERVER_HISTOGRAM_SIZE_BUCKETS")
	if len(strBuckets) > 0 {
		ret.BucketBytesServer = loadBuckets(strBuckets)
	}
	if len(ret.BucketBytesServer) == 0 {
		ret.BucketBytesServer = DefaultSizeBuckets
	}
	log.Infof("SERVER_HISTOGRAM_SIZE_BUCKETS=%v", ret.BucketBytesServer)

	strBuckets = os.Getenv("SERVER_HISTOGRAM_DURATION_BUCKETS")
	if len(strBuckets) > 0 {
		ret.BucketSecondsServer = loadBuckets(strBuckets)
	}
	if len(ret.BucketSecondsServer) == 0 {
		ret.BucketSecondsServer = DefaultTimeBuckets
	}
	log.Infof("SERVER_HISTOGRAM_DURATION_BUCKETS=%v", ret.BucketSecondsServer)

	strBuckets = os.Getenv("CLIENT_HISTOGRAM_SIZE_BUCKETS")
	if len(strBuckets) > 0 {
		ret.BucketBytesClient = loadBuckets(strBuckets)
	}
	if len(ret.BucketBytesClient) == 0 {
		ret.BucketBytesClient = DefaultSizeBuckets
	}
	log.Infof("CLIENT_HISTOGRAM_SIZE_BUCKETS=%v", ret.BucketBytesClient)

	strBuckets = os.Getenv("CLIENT_HISTOGRAM_DURATION_BUCKETS")
	if len(strBuckets) > 0 {
		ret.BucketSecondsClient = loadBuckets(strBuckets)
	}
	if len(ret.BucketSecondsClient) == 0 {
		ret.BucketSecondsClient = DefaultTimeBuckets
	}
	log.Infof("CLIENT_HISTOGRAM_DURATION_BUCKETS=%v", ret.BucketSecondsClient)

	return ret
}

// loadBuckets takes in a bucket array in string representation (i.e. 'strBuckets')
// and converts it to []float64. In case of error, it is logged and return nil.
func loadBuckets(strBuckets string) []float64 {
	aryBuckets := strings.Split(strBuckets, ",")
	if len(aryBuckets) == 0 {
		log.WithField("buckets_str", strBuckets).Error("err parse buckets: failed to split intervals, ignore it")
		return nil
	}
	buckets := make([]float64, len(aryBuckets))
	for i, strInterval := range aryBuckets {
		strInterval = strings.TrimSpace(strInterval)
		interval, err := strconv.ParseFloat(strInterval, 64)
		if err != nil {
			log.WithFields(logrus.Fields{
				"buckets_str": strBuckets,
				"err":         err,
			}).Errorf("err convert interval [%s] to float", strInterval)
			return nil
		}
		buckets[i] = interval
	}
	return buckets
}

// Register registers one or more Vectors.
func Register(vecs ...*Vector) {
	log.WithField("vecs", vecs).Debug("@metrics.Register")
	for _, vec := range vecs {
		if _, found := Vectors[vec.Name]; found {
			log.WithField("Name", vec.Name).Panic("Vector Name already exists")
		}
		switch vec.Type {
		case TypeCounter, TypeGauge, TypeSummary, TypeHistogram: // Valid, do nothing.
		default:
			log.WithField("Type", vec.Type).Panic("Invalid Vector Type")
		}
		if vec.Desc == "" {
			log.WithField("Name", vec.Name).Panic("Vector Desc is required")
		}
		Vectors[vec.Name] = vec
	}
}

// Serve sets up and serves with its Adapter.
func Serve() {

	// load caller specified buckets from ENV (iff any)
	customBuckets := newHistogramBuckets()

	// Web server metrics.
	labels := []string{"route", "addr", "method", "status"}
	Register(
		&Vector{
			Type:   TypeCounter,
			Name:   "http_request_total",
			Desc:   "HTTP total server requests",
			Labels: labels,
		},
		&Vector{
			Type:    TypeHistogram,
			Name:    "http_request_bytes",
			Desc:    "HTTP server request data in bytes",
			Labels:  labels,
			Buckets: customBuckets.BucketBytesServer,
		},
		&Vector{
			Type:    TypeHistogram,
			Name:    "http_request_duration_seconds",
			Desc:    "HTTP server request time in seconds",
			Labels:  labels,
			Buckets: customBuckets.BucketSecondsServer,
		},
	)

	// Web client metrics.
	labels = []string{"url", "method", "status"}
	Register(
		&Vector{
			Type:   TypeCounter,
			Name:   "client_http_request_total",
			Desc:   "HTTP total client requests",
			Labels: labels,
		},
		&Vector{
			Type:    TypeHistogram,
			Name:    "client_http_request_bytes",
			Desc:    "HTTP client request data in bytes",
			Labels:  labels,
			Buckets: customBuckets.BucketBytesClient,
		},
		&Vector{
			Type:    TypeHistogram,
			Name:    "client_http_request_duration_seconds",
			Desc:    "HTTP client request time in seconds",
			Labels:  labels,
			Buckets: customBuckets.BucketSecondsClient,
		},
	)

	// Health metrics.
	labels = []string{"dep", "state"} // "addr" ?.
	Register(
		&Vector{
			Type:   TypeCounter,
			Name:   "health_check_total",
			Desc:   "Health total checks",
			Labels: labels,
		},
		&Vector{
			Type:   TypeHistogram,
			Name:   "health_check_duration_ms",
			Desc:   "Health check time in milliseconds",
			Labels: labels,
		},
	)

	// Service metrics.
	Register(
		&Vector{
			Type:   TypeGauge,
			Name:   "version",
			Desc:   "Service version",
			Labels: []string{"version", "git_hash", "build_stamp"},
		},
	)

	Adapter.Serve()

	Served = true

	Set("version", 1, sys.Version, sys.GitHash, sys.BuildStamp)
}

// WebHandler provides web handler.
func WebHandler() http.Handler {
	return Adapter.WebHandler()
}

func do(reset bool, name string, val float64, labels ...string) {
	logger := log.WithFields(logrus.Fields{"Reset": reset, "Name": name, "Val": val, "Labels": labels})
	logger.Debug("@metrics.Do")

	if !Served {
		logger.Info("skipped metric")
		return
	}

	Stats.Total++
	vec, found := Vectors[name]
	if !found {
		log.WithField("Name", name).Panic("Vector not found")
	}
	if reset && vec.Type != TypeGauge {
		log.WithField("Type", vec.Type).Panic("Reset not allowed for Vector Type")
	}
	Adapter.Do(reset, vec, val, labels...)
}

// Add provides polymorphism for [Inc, Dec, Add, Sub, Observe] regardless of Type.
// It increments (Counter or Gauge), decrements (Gauge), or observes (Summary, Histogram) by arbitrary value.
// Requires explicit val, even for 1 or -1.
func Add(name string, val float64, labels ...string) {
	do(false, name, val, labels...)
}

// Set sets Gauge to arbitrary value.
func Set(name string, val float64, labels ...string) {
	do(true, name, val, labels...)
}
