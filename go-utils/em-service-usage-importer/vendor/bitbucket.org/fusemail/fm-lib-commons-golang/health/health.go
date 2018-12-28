/*
Package health provides health with dependencies.
Duties:
- Provides Depender interface to implement concrete dependencies via Dependency struct.
- Provides Register to register one or more dependencies.
- Runs checker infinite loop (every CheckInterval in StartChecker) in the background,
  to update Health status for all concrete dependencies.
- Uses SyncMap as concurrent maps.
- Provides WebHandler to expose web endpoint.
- Sets up everything via Serve.
- Instruments metrics:
	health_check_total, health_check_duration_ms.
- Generates statistics via Stats.
- Logs pertinent info.
See deps for commons concrete Depender examples.
*/
package health

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"bitbucket.org/fusemail/fm-lib-commons-golang/metrics"
	"bitbucket.org/fusemail/fm-lib-commons-golang/sys"
	"bitbucket.org/fusemail/fm-lib-commons-golang/utils"
	"github.com/sirupsen/logrus"
)

const (
	// StatusHealthy defines the status code for a healthy state.
	StatusHealthy = http.StatusOK

	stateHealthy   = "healthy"
	stateUnhealthy = "unhealthy"
)

var (
	// Served indicates whether health has been served.
	Served bool

	// Config holds the health configuration.
	Config = &struct {
		StatusUnhealthy int           `json:"status_unhealthy"`  // Status code for an unhealthy state (at least one dependency with error).
		CheckInterval   time.Duration `json:"check_interval"`    // How often dependencies must be checked.
		CheckMaxTimeout time.Duration `json:"check_max_timeout"` // Maximum timeout for each dependency check.

		LogChecks               bool          `json:"log_checks"`                // Log check infos.
		MinimumCheckInterval    time.Duration `json:"min_check_interval"`        // Minimum duration to wait between health checks.
		CheckIntervalSubtrahend time.Duration `json:"check_interval_subtrahend"` // Time to subtract from CheckInterval in order to apply timeouts.
	}{
		StatusUnhealthy: http.StatusServiceUnavailable,
		CheckInterval:   15 * time.Second,
		CheckMaxTimeout: 14 * time.Second,

		MinimumCheckInterval:    2 * time.Second,
		CheckIntervalSubtrahend: 500 * time.Millisecond,
	}

	// Stats contains health statistics.
	Stats = struct {
		Total uint64 `json:"total"`
		Fails uint64 `json:"fails"`

		TotalRequests uint64 `json:"total_requests"` // @ WebHandler.
		TotalChecks   uint64 `json:"total_checks"`   // @ StartChecker's loop.

		CheckDurationMS int64 `json:"check_duration_ms"`
	}{}

	// Dependencies holds all registered concrete dependencies.
	Dependencies = map[string]*Dependency{}

	// Health holds the status of all checked dependencies.
	Health = struct {
		Version      map[string]string `json:"version"`      // Map for version/build info.
		Dependencies *utils.SyncMap    `json:"dependencies"` // Map for all dependencies.
		Health       *utils.SyncMap    `json:"health"`       // Map for single health state.
	}{
		Dependencies: utils.NewSyncMap(),
		Health:       utils.NewSyncMap(),
	}

	// Errors.
	errUnhealthyDefault  = errors.New("starting (unhealthy by default)")
	errCheckerNotStarted = errors.New("checker NOT yet started")

	// Error messages.
	errMsgUnhealthy     = "Unhealthy"
	errMsgFailedMarshal = "Failed to marshal Health"
	errMsgFailedWrite   = "Failed to write Health response"
	errMsgCheckTimeout  = "Health dependency check has timed out after %v"

	// Package logger, set with SetLogger.
	log = logrus.StandardLogger()
)

// SetLogger overrides the package logger.
func SetLogger(logger *logrus.Logger) {
	log = logger
}

// Depender defines the interface for all concrete dependency implementations.
type Depender interface {
	Check() (map[string]interface{}, error) // Checks health, expects optional config/state map, and and error (nil if healthy).
}

// Dependency defines a registered dependency.
type Dependency struct {
	Name string   `json:"-"`
	Desc string   `json:"desc"`
	Item Depender `json:"item"`
	Key  string   `json:"name"` // Unique, as lowercase Name.
}

func (d *Dependency) String() string {
	return fmt.Sprintf(
		"{%T: %v (%v) @ %v}",
		d, d.Name, d.Desc, d.Item,
	)
}

// Register registers one or more Dependencies.
func Register(dependencies ...*Dependency) {
	log.WithField("dependencies", dependencies).Debug("@health.Register")

	for _, dependency := range dependencies {
		// Validate Name as required.
		if dependency.Name == "" {
			log.WithField("dependency", dependency).Panic("Dependecy's Name is required")
		}

		// Validate Item as required.
		if dependency.Item == nil {
			log.WithField("dependency", dependency).Panic("Dependecy's Item is required")
		}

		// Validate dependency.Name as unique.
		dependency.Key = strings.ToLower(dependency.Name)

		if _, found := Dependencies[dependency.Key]; found {
			log.WithFields(logrus.Fields{"dependency": dependency, "key": dependency.Key}).Panic("Dependencies must be unique by Name")
		}

		// Validate marshal on Register, in order to detect errors (and panic) at early stage.
		// For reference, error if "chan bool" struct field is defined in a dependency:
		//   json: unsupported type: chan bool
		// Fix any errors by defining the necessary methods or struct tags, e.g.
		//   Field TYPE `json:"-"`
		// See IgnoreMe field in Sample struct @ sample.go.
		_, err := json.Marshal(dependency.Item)

		if err != nil {
			log.WithFields(logrus.Fields{"dependency": dependency, "err": err}).Panic("Failed to marshal Dependency's Item (Depender)")
		}

		Dependencies[dependency.Key] = dependency
	}
}

type depCheck struct {
	dependency     *Dependency
	durationMillis int64
	state          utils.OpenMap
	err            error
}

func setDep(dc depCheck) {
	log.WithField("dc", dc).Debug("@setDep")

	dv := utils.OpenMap{
		"dependency":  dc.dependency,
		"duration_ms": dc.durationMillis,
	}

	if dc.state != nil {
		dv["state"] = dc.state
	}

	hstate := stateHealthy

	ready := dc.err == nil
	dv["ready"] = ready
	atomic.AddUint64(&Stats.Total, 1)

	if !ready {
		atomic.AddUint64(&Stats.Fails, 1)
		dv["error"] = dc.err.Error()

		if dc.err != errUnhealthyDefault {
			log.WithField("dependency", dv).Error("unhealthy dependency")
		}

		hstate = stateUnhealthy
	}

	log.WithField("dv", dv).Debug("@setDep")
	Health.Dependencies.Store(dc.dependency.Key, dv)

	// Health metrics.
	// Skip metrics for unhealthy default and if not yet served.
	if dc.err == errUnhealthyDefault || !metrics.Served {
		return
	}

	labels := []string{
		dc.dependency.Key,
		hstate,
	}

	metrics.Add(
		"health_check_total",
		1,
		labels...,
	)
	metrics.Add(
		"health_check_duration_ms",
		float64(dc.durationMillis),
		labels...,
	)
}

// StartChecker loops every interval (CheckInterval) to update status of dependencies.
// TODO - complete MAIL-912.
func StartChecker() {
	log.Debug("@StartChecker")

	// Validate minimum interval.
	if Config.CheckInterval < Config.MinimumCheckInterval {
		log.WithFields(logrus.Fields{"interval": Config.CheckInterval, "min": Config.MinimumCheckInterval}).Panic("Invalid interval - too short")
	}

	_set := func(key string, val interface{}) {
		log.WithFields(logrus.Fields{"key": key, "val": val}).Debug("@StartChecker._set")
		Health.Health.Store(key, val)
	}

	// Initialize all as unhealthy.
	for _, dependency := range Dependencies {
		setDep(depCheck{
			dependency: dependency,
			err:        errUnhealthyDefault,
		})
	}

	// Started must be set AFTER initialization above,
	// used for overall healthy status in WebHandler.
	started := time.Now()
	_set("started", started)

	timeout := Config.CheckInterval - Config.CheckIntervalSubtrahend
	if timeout > Config.CheckMaxTimeout {
		timeout = Config.CheckMaxTimeout
	}
	_set("config_interval", utils.StrV(Config.CheckInterval))
	_set("config_timeout", utils.StrV(timeout))

	log.WithFields(logrus.Fields{"interval": Config.CheckInterval, "timeout": timeout, "started": started}).Info("starting health checker")

	// Infinite loop.
	for {
		atomic.AddUint64(&Stats.TotalChecks, 1)
		log.WithField("interval", Config.CheckInterval).Debug("checking health")

		for _, dependency := range Dependencies {
			go func(dependency *Dependency) {
				chChecked := make(chan int64, 1) // buffer=1 to avoid goroutine leak

				go func(dependency *Dependency) {
					dependencyStart := time.Now()
					state, err := dependency.Item.Check()
					ms := utils.ElapsedMillis(dependencyStart)
					setDep(depCheck{
						dependency:     dependency,
						durationMillis: ms,
						state:          state,
						err:            err,
					})
					chChecked <- ms
				}(dependency)

				// Watch timeout.
				select {
				case ms := <-chChecked:
					if Config.LogChecks {
						log.WithFields(logrus.Fields{"dependency": dependency, "duration_ms": ms}).Info("health dependency check completed")
					}
				case <-time.After(timeout):
					emsg := fmt.Sprintf(errMsgCheckTimeout, timeout)
					log.WithFields(logrus.Fields{"dependency": dependency, "timeout": timeout}).Warn(emsg)
					setDep(depCheck{
						dependency:     dependency,
						durationMillis: utils.DurationMillis(timeout),
						err:            errors.New(emsg),
					})
				}
			}(dependency)
		}

		last := time.Now()
		_set("last", last)

		Stats.CheckDurationMS = utils.ElapsedMillis(started, last)
		_set("duration_ms", Stats.CheckDurationMS)

		log.WithField("interval", Config.CheckInterval).Debug("Sleeping health")
		time.Sleep(Config.CheckInterval)
	}
}

// WebHandler provides web handler.
func WebHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&Stats.TotalRequests, 1)

		var hstatus int // Header status to be written.

		hmap := Health.Health

		_setStatus := func(status int) {
			hstate := stateHealthy
			ready := status == StatusHealthy
			hmap.Store("ready", ready)
			if !ready {
				// Do NOT WriteHeader here, first check other potential errors (e.g. json marshal).
				hstatus = status
				hstate = stateUnhealthy
			}
			hmap.Store("status", status)
			hmap.Store("state", hstate)
		}

		// Must return after _err call.
		_err := func(er error, msg string) {
			log.WithFields(logrus.Fields{"err": er, "msg": msg, "health": Health}).Error("error during health web handler")
			errStatus := http.StatusInternalServerError
			_setStatus(errStatus)
			w.WriteHeader(errStatus)
			fmt.Fprint(w, msg, "\n", er.Error())
		}

		// Unhealthy if not yet started.
		if _, found := hmap.Load("started"); !found {
			_err(errCheckerNotStarted, errMsgUnhealthy)
			return
		}

		// Unhealthy if any dependency contains error.
		healthy := true

		Health.Dependencies.Range(func(_, d interface{}) bool {
			hdep := d.(utils.OpenMap)

			if _, found := hdep["error"]; found {
				// Even if unhealthy, do NOT fail and return, but instead
				// let it generate the usual json contents BUT with unhealthy header.
				log.WithFields(logrus.Fields{"dependency": hdep}).Info("unhealthy dependencies (breaking on first)")
				_setStatus(Config.StatusUnhealthy)
				healthy = false
				return false
			}

			return true
		})

		if healthy {
			_setStatus(StatusHealthy)
		}

		hbyts, err := json.Marshal(Health)

		hmap.Delete("status")
		hmap.Delete("state")

		if err != nil {
			_err(err, errMsgFailedMarshal)
			return
		}

		// No marshal errors, so write this header BEFORE WriteHeader below.
		w.Header().Set("Content-Type", "application/json")

		// Write status header AFTER json marshal check above,
		// in order to let it write the correct header status (errStatus) first,
		// instead of a potential StatusUnhealthy (which can't be overriden once written).
		// Note that 200 OK status does not require explicit write.
		if hstatus != 0 {
			w.WriteHeader(hstatus)
		}

		_, err = w.Write(hbyts)
		if err != nil {
			_err(err, errMsgFailedWrite)
			return
		}
	})
}

// Serve sets up and serves, forking checker.
func Serve() {
	Health.Version = sys.BuildInfo

	if len(Dependencies) == 0 {
		log.Warn("no health dependencies detected, use health.Register")
	}

	go StartChecker()

	Served = true
}
