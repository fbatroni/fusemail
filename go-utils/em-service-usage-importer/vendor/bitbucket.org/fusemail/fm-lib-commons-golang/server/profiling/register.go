package profiling

import (
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
)

// RegisterMux - only works with gorilla.
func RegisterMux(router *mux.Router) {
	// add the profiling endpoints
	router.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	router.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	router.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	router.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	router.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))
	router.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
}
