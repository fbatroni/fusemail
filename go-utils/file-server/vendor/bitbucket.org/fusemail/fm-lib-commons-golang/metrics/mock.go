package metrics

// Provides Mock Adapter.

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// MockAdapter must satisfy Adaptor interface.
type MockAdapter struct {
	Account map[string]float64
}

func (a *MockAdapter) String() string {
	return fmt.Sprintf("{%T}", a)
}

// Serve sets up and serves the Adapter.
func (a *MockAdapter) Serve() {
	a.Account = make(map[string]float64)
	log.WithField("Adapter", a).Debug("@MockAdapter.Serve")
}

// WebHandler returns the web handler.
func (a *MockAdapter) WebHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "MockAdapter Metrics web handler \n %+v", a.Account)
	})
}

// Do executes the Add or Set operation.
func (a *MockAdapter) Do(reset bool, vec *Vector, val float64, labels ...string) {
	log.WithFields(logrus.Fields{"Adapter": a, "Reset": reset, "Vector": vec, "Val": val, "Labels": labels}).Debug("@MockAdapter.Do")
	kn := append([]string{vec.Name}, labels...)
	k := strings.Join(kn, "~")
	v := val
	if !reset {
		v += a.Account[k]
	}
	a.Account[k] = v
	log.WithFields(logrus.Fields{"Adapter": a, "k": k, "v": v}).Debug("@MockAdapter.Do (end)")
}
