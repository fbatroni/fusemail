package metrics

// Provides Prometheus Adapter.

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusAdapter must satisfy Adaptor interface.
type PrometheusAdapter struct {
}

func (a *PrometheusAdapter) String() string {
	return fmt.Sprintf("{%T}", a)
}

// Serve sets up and serves the Adapter.
func (a *PrometheusAdapter) Serve() {
	log.WithField("Adapter", a).Debug("@PrometheusAdapter.Serve")
	for _, vec := range Vectors {
		switch vec.Type {
		case TypeCounter:
			v := prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: vec.Name,
					Help: vec.Desc,
				},
				vec.Labels,
			)
			prometheus.MustRegister(v)
			vec.Ref = v
		case TypeGauge:
			v := prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: vec.Name,
					Help: vec.Desc,
				},
				vec.Labels,
			)
			prometheus.MustRegister(v)
			vec.Ref = v
		case TypeSummary:
			v := prometheus.NewSummaryVec(
				prometheus.SummaryOpts{
					Name: vec.Name,
					Help: vec.Desc,
				},
				vec.Labels,
			)
			prometheus.MustRegister(v)
			vec.Ref = v
		case TypeHistogram:
			v := prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    vec.Name,
					Help:    vec.Desc,
					Buckets: vec.Buckets,
				},
				vec.Labels,
			)
			prometheus.MustRegister(v)
			vec.Ref = v
		default:
			log.WithFields(logrus.Fields{"Adapter": a, "Vector": vec, "Type": vec.Type}).Panic("Invalid Vector Type")
		}
	}
}

// WebHandler returns the web handler.
func (a *PrometheusAdapter) WebHandler() http.Handler {
	return promhttp.Handler()
}

// Do executes the Add or Set operation.
func (a *PrometheusAdapter) Do(reset bool, vec *Vector, val float64, labels ...string) {
	log.WithFields(logrus.Fields{"Adapter": a, "Reset": reset, "Vector": vec, "Val": val, "Labels": labels}).Debug("@PrometheusAdapter.Do")
	switch v := vec.Ref.(type) {
	case *prometheus.CounterVec:
		v.WithLabelValues(labels...).Add(val)
	case *prometheus.GaugeVec:
		m := v.WithLabelValues(labels...)
		if reset {
			m.Set(val)
		} else {
			m.Add(val)
		}
	case *prometheus.SummaryVec:
		v.WithLabelValues(labels...).Observe(val)
	case *prometheus.HistogramVec:
		v.WithLabelValues(labels...).Observe(val)
	default:
		log.WithFields(logrus.Fields{"Adapter": a, "Vector": vec, "Ref": vec.Ref}).Panic("Invalid Prometheus Ref")
	}
}
