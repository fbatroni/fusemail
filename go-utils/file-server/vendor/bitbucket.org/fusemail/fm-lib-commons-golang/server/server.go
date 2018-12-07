/*
Package server provides web server capabilities.
Duties:
- Provides ApplicationOptions for default server options.
- Provides ServerConfig to configure web servers, with full setup and serve.
- Supports both HTTP and HTTPS (TLS/SSL) via ServerConfig struct.
- Holds Defaults for new servers.
- Exposes WriteJSON to write data as JSON, with proper header, logging and error handling.
- Contains middleware sub-package.
- Contains handles sub-package.
- Generates statistics via Stats.
- Logs pertinent info, including LogRequests.
Router examples (standard, chi, gorilla)  @  /fm-service-DEMO/extra.go
*/
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"bitbucket.org/fusemail/fm-lib-commons-golang/utils"
	"github.com/sirupsen/logrus"
)

var (
	// Stats contains web server statistics.
	Stats = struct {
		Total   int `json:"total"`
		Curr    int `json:"curr"`
		Rejects int `json:"rejects"`
	}{}

	// Defaults holds the default values for Options.
	Defaults = &Config{
		Port:            8080,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 60 * time.Second,
	}

	// ShutdownAllDefaultTimeout indicates the default ShutdownAll timeout.
	ShutdownAllDefaultTimeout = 60 * time.Second

	// shuttingDown indicates whether servers are being shutdown.
	shuttingDown bool

	// Servers keeps track of all live web servers via port keys.
	Servers = make(map[int]*Server)

	// Package logger, set with SetLogger.
	log = logrus.StandardLogger()
)

// SetLogger overrides the package logger.
func SetLogger(logger *logrus.Logger) {
	log = logger
}

// ShutdownErr collects all server errors during shutdown.
type ShutdownErr struct {
	ErrList []error
}

func (s *ShutdownErr) Error() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("failed to shutdown server(s) [%d]", len(s.ErrList)))
	for _, err := range s.ErrList {
		buf.WriteString(": " + err.Error())
	}
	return buf.String()
}

// Config provides the structure to setup web servers, either HTTP or HTTPS (SSL/TLS).
type Config struct {
	Port   int  `json:"port"`
	UseSSL bool `json:"use_ssl"`

	// Required only if UseSSL.
	SSLCert string `json:"ssl_cert"`
	SSLKey  string `json:"ssl_key"`

	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

	Router http.Handler `json:"-"`

	// Server to use as fallback if Port is zero on Setup.
	Fallback *Server `json:"fallback"`
}

// NewConfig constructs server config instances.
func NewConfig() *Config {
	s := &Config{
		Port:            Defaults.Port,
		ReadTimeout:     Defaults.ReadTimeout,
		WriteTimeout:    Defaults.WriteTimeout,
		ShutdownTimeout: Defaults.ShutdownTimeout,
	}
	return s
}

// Server holds a live server with corresponding setup config (including router).
type Server struct {
	Started time.Time    `json:"started"`
	Config  *Config      `json:"config"`
	Server  *http.Server `json:"-"`
}

func (s *Server) String() string {
	return fmt.Sprintf("{%T: %+v}", s, s.Config)
}

/*
Shutdown tries to gracefully shutdown the enclosed http server.
If Config.ShutdownTimeout is non-zero, use it with ctx,
otherwise (zero config timeout) use ctx with its embedded timeout (if any).
Returns error from http server shutdown.
*/
func (s *Server) Shutdown(ctx context.Context) error {
	log.WithField("server", s).Debug("shutdown web server")

	if s.Config.ShutdownTimeout != 0 {
		var cancelFunc context.CancelFunc

		// Add local timeout to context.
		ctx, cancelFunc = context.WithTimeout(ctx, s.Config.ShutdownTimeout)
		defer cancelFunc()
	}

	return s.Server.Shutdown(ctx)
}

/*
Setup sets up a (potentially new) web server based on config.
(Re)uses fallback if provided (not nil) and if port is zero.
If another server is detected with the same port, it will reuse it (instead of creating or failing).
In case of server reuse, new config are not applied.
Return (new server, true) if new port, otherwise return (reusable server, false) if reusing port.
*/
func Setup(config *Config) (*Server, bool) {
	log.WithField("config", config).Debug("@server.Setup")

	if config == nil || config.Port == 0 {
		if config.Fallback != nil {
			return config.Fallback, false
		}

		config = NewConfig()
	}

	if s, found := Servers[config.Port]; found {
		log.WithField("server", s).Info("reusing web server")
		return s, false
	}

	s := &Server{
		Server: &http.Server{
			Addr:         fmt.Sprintf(":%d", config.Port),
			Handler:      config.Router,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
		},
		Config: config,
	}

	log.WithField("server", s).Info("creating web server")
	Servers[config.Port] = s

	return s, true
}

/*
Serve starts all web servers.
It does NOT automatically block (anymore),
but instead lets the app choose the pertinent blocking mechanism (e.g. channels).
Alternatively you can use server.BlockAndExit.
*/
func Serve() {
	log.WithField("servers", Servers).Info("servers")

	// Fork start web servers.
	for _, s := range Servers {
		s.Started = time.Now()

		go func(s *Server) {
			var err error

			logger := log.WithField("server", s)
			logger.Info("starting web server")

			if s.Config.UseSSL {
				err = s.Server.ListenAndServeTLS(s.Config.SSLCert, s.Config.SSLKey)
			} else {
				err = s.Server.ListenAndServe()
			}

			logger = logger.WithField("err", err)
			if !shuttingDown && err != http.ErrServerClosed {
				logger.Fatal("failed to start web server")
			}
			logger.Info("graceful web server shutdown")
		}(s)
	}
}

// DataToJSON takes some data and tries to marshal it to JSON for a HTTP response
func DataToJSON(data interface{}) ([]byte, error) {
	if data == nil {
		data = utils.OpenMap{}
	}
	d, err := json.Marshal(data)
	if err != nil {
		d = ErrorToJSON(err)
	}
	return d, err
}

// ErrorToJSON marshals an error into some JSON for a HTTP response
func ErrorToJSON(err error) []byte {
	d, _ := json.Marshal(map[string]string{"error": err.Error()})
	return d
}

// WriteJSON writes data into w, with proper header, logging, and error handling.
func WriteJSON(w http.ResponseWriter, data interface{}) {
	WriteJSONWithStatus(w, data, http.StatusOK)
}

// WriteJSONWithStatus writes the data as JSON with the provided http status
func WriteJSONWithStatus(w http.ResponseWriter, data interface{}, status int) {
	d, err := DataToJSON(data)
	if err != nil {
		writeWithStatus(w, ErrorToJSON(err), http.StatusInternalServerError)
		return
	}
	writeWithStatus(w, d, status)
}

// WriteJSONErrorWithStatus writes the error to the response with the provided HTTP status code
func WriteJSONErrorWithStatus(w http.ResponseWriter, err error, status int) {
	d, err := json.Marshal(map[string]string{"error": err.Error()})
	if err != nil {
		writeWithStatus(w, []byte("marshal json error"), http.StatusInternalServerError)
		return
	}
	writeWithStatus(w, d, status)
}

func writeWithStatus(w http.ResponseWriter, data []byte, status int) {
	log.WithField("data", string(data)).Debug("@WriteJSON")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := w.Write(data)
	if err != nil {
		log.WithField("error", err).Error("failed to write response")
	}
}

/*
ShutdownAll gracefully shuts down all web servers, with default timeout.
Returns ShutdownErr containing a list of server errors.
*/
func ShutdownAll() error {
	log.Debug("shutdown all web servers with default timeout")
	return ShutdownAllWithTimeout(ShutdownAllDefaultTimeout)
}

/*
ShutdownAllWithTimeout gracefully shuts down all web servers, according to maxTimeout.
Non-zero maxTimeout indicates maximum time to wait for all servers to shutdown,
otherwise (zero maxTimeout) per-server shutdown timeout is used (i.e. Config.ShutdownTimeout).
If per-server timeout is zero, each server's shutdown will wait indefinitely
(until all live connections become idle and closed).
Returns ShutdownErr containing a list of server errors.
*/
func ShutdownAllWithTimeout(maxTimeout time.Duration) error {
	log.WithField("timeout", maxTimeout).Info("shutdown all web servers with max timeout")

	var cancelFunc context.CancelFunc
	ctx := context.Background()

	if maxTimeout != 0 {
		ctx, cancelFunc = context.WithTimeout(ctx, maxTimeout)
		defer cancelFunc()
	}

	return ShutdownAllWithContext(ctx)
}

/*
ShutdownAllWithContext gracefully shuts down all web servers, with context.
Returns ShutdownErr containing a list of server errors.
*/
func ShutdownAllWithContext(ctx context.Context) error {
	log.WithField("ctx", ctx).Info("shutdown all web servers with context")

	var (
		wg   sync.WaitGroup
		errs []error
	)

	shuttingDown = true

	for _, s := range Servers {
		wg.Add(1)

		go func(svr *Server) {
			if err := svr.Shutdown(ctx); err != nil {
				log.WithFields(logrus.Fields{"server": svr, "err": err}).Error("failed to shutdown web server")
				errs = append(errs, err)
			}

			wg.Done()
		}(s)
	}

	wg.Wait()

	shuttingDown = false

	if len(errs) > 0 {
		return &ShutdownErr{ErrList: errs}
	}

	return nil
}
