package middleware

import (
	"fmt"
	"net/http"
	"time"

	"bitbucket.org/fusemail/fm-lib-commons-golang/metrics"
	"github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

// Metrics is a middleware handler that applies metrics to requests.
type Metrics struct{}

// NewMetrics constructs metrics instances.
func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()

	next(rw, r)

	if !metrics.Served {
		return
	}

	res := rw.(negroni.ResponseWriter)

	/*
		With respect to "route" label value for http_request_* metrics,
		use current url path by default,
		unless we can extract the route template (hardcoded to use mux for now),
		specifically in the case of dynamic paths in order to minimize metrics cardinality.
	*/
	reqPath := r.URL.Path
	route := mux.CurrentRoute(r)
	if route != nil {
		tpl, err := route.GetPathTemplate()
		if err != nil {
			GetContextLogger(r.Context()).WithFields(logrus.Fields{
				"reqPath": reqPath, "tpl": tpl, "err": err,
			}).Warn("failed to get path template in middleware metrics")
		} else {
			reqPath = tpl
		}
	}

	labels := []string{
		reqPath,
		r.Host,
		r.Method,
		fmt.Sprintf("%d", res.Status()),
	}

	metrics.Add(
		"http_request_total",
		1,
		labels...,
	)
	metrics.Add(
		"http_request_bytes",
		float64(res.Size()),
		labels...,
	)
	metrics.Add(
		"http_request_duration_seconds",
		time.Since(start).Seconds(),
		labels...,
	)
}
