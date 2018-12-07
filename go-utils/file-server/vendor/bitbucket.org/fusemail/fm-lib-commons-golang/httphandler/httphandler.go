package httphandler

import (
	"bitbucket.org/fusemail/fm-lib-commons-golang/bindata"
	"bitbucket.org/fusemail/fm-lib-commons-golang/metrics"
	"bitbucket.org/fusemail/fm-lib-commons-golang/server"
	"bitbucket.org/fusemail/fm-lib-commons-golang/server/handlers"
	"bitbucket.org/fusemail/fm-lib-commons-golang/server/middleware"
	"bitbucket.org/fusemail/fm-lib-commons-golang/server/profiling"
	log "github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

// RequestTrace contains request details to apply trace function to.
type RequestTrace struct {
	Logger *log.Entry
	Action string
	Error  error
}

// RequestTraceFunc is a closure function that implements on a RequestTracer instance. e.g. reportMetric, logging etc.
type RequestTraceFunc func(*RequestTrace)

// ReportErrorMetric returns a request trace function that send error counter metric on trace.
func ReportErrorMetric(m *metrics.Metric) RequestTraceFunc {
	return func(t *RequestTrace) {
		if t.Error != nil {
			m.AddOne(t.Action, t.Error.Error())
		}
	}
}

// LogError returns a request trace func that logs error on trace.
func LogError() RequestTraceFunc {
	return func(t *RequestTrace) {
		if t.Error != nil {
			t.Logger.WithField("error", t.Error).Error(t.Action)
		}
	}
}

// HTTPHandler serves HTTP endpoints
type HTTPHandler struct {
	Router           *mux.Router
	CommonMiddleware *negroni.Negroni
	RequestLimit     int
	traceFunc        []RequestTraceFunc
}

// New initiates a HTTPHandler instance.
func New(router *mux.Router, common *negroni.Negroni, limit int) *HTTPHandler {
	return &HTTPHandler{
		Router:           router,
		CommonMiddleware: common,
		RequestLimit:     limit,
		traceFunc: []RequestTraceFunc{
			ReportErrorMetric(metrics.ErrorCounter),
			LogError(),
		},
	}
}

// Limiter returns middleware that limits requests on RequestLimit
func (h *HTTPHandler) Limiter() *negroni.Negroni {
	return h.CommonMiddleware.With(middleware.NewLimiter(h.RequestLimit))
}

// MountDefaultEndpoints mounts applications default endpoints for health, metrics, docs, log, and trace.
func (h *HTTPHandler) MountDefaultEndpoints(options server.ApplicationOptions) {
	router := h.Router
	common := h.CommonMiddleware

	router.Handle(options.DocRoute,
		common.With(negroni.Wrap(handlers.DocsMarkdown(bindata.LoadFile, options.DocFile))))
	router.Handle(options.APIRoute,
		common.With(negroni.Wrap(handlers.DocsHTML(bindata.LoadFile, options.APIFile))))
	router.Handle(options.HealthRoute,
		common.With(negroni.Wrap(handlers.Health())))
	router.Handle(options.MetricsRoute,
		common.With(negroni.Wrap(handlers.Metrics())))
	router.Handle(options.VersionRoute,
		common.With(negroni.Wrap(handlers.Version())))
	router.Handle(options.SysRoute,
		common.With(negroni.Wrap(handlers.Sys(options))))
	router.Handle(options.LogRoute,
		common.With(negroni.Wrap(handlers.Log(options.LogCount))))
	router.Handle(options.TraceRoute,
		common.With(negroni.Wrap(handlers.Trace())))

	// Mount profiler.
	profiling.RegisterMux(router)
}

// AddTracer appends a new trace function to existing tracers list.
func (h *HTTPHandler) AddTracer(tracer RequestTraceFunc) {
	h.traceFunc = append(h.traceFunc, tracer)
}

// Trace implements all trace functions to given trace details.
func (h *HTTPHandler) Trace(t *RequestTrace) {
	for _, f := range h.traceFunc {
		f(t)
	}
}
