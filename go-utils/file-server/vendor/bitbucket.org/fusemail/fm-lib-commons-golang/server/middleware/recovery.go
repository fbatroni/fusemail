package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
)

// Recovery is a middleware handler that recovers from any panics and writes a 500 if there was one.
type Recovery struct {
	PrintStack       bool
	ErrorHandlerFunc func(interface{})
	StackAll         bool
	StackSize        int
}

// NewRecovery constructs recovery instances.
func NewRecovery() *Recovery {
	return &Recovery{
		PrintStack: true,
		StackAll:   false,
		StackSize:  1024 * 8,
	}
}

func (rec *Recovery) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			if rw.Header().Get("Content-Type") == "" {
				rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			}

			rw.WriteHeader(http.StatusInternalServerError)

			stack := make([]byte, rec.StackSize)
			stack = stack[:runtime.Stack(stack, rec.StackAll)]

			f := "PANIC: %s stack: %s"
			log.Errorf(f, err, stack)

			if rec.PrintStack {
				fmt.Fprintf(rw, f, err, stack)
			}

			if rec.ErrorHandlerFunc != nil {
				func() {
					defer func() {
						if err = recover(); err != nil {
							log.Errorf("provided ErrorHandlerFunc panic'd: %s, trace: %s", err, debug.Stack())
							log.Errorf("%s", debug.Stack())
						}
					}()
					rec.ErrorHandlerFunc(err)
				}()
			}
		}
	}()

	next(rw, r)
}
