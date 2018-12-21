package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/satori/go.uuid"

	"bitbucket.org/fusemail/fm-lib-commons-golang/deps"
	"bitbucket.org/fusemail/fm-lib-commons-golang/health"
	"bitbucket.org/fusemail/fm-lib-commons-golang/httphandler"
	"bitbucket.org/fusemail/fm-lib-commons-golang/metrics"
	"bitbucket.org/fusemail/fm-lib-commons-golang/server"
	"bitbucket.org/fusemail/fm-lib-commons-golang/server/middleware"
	"bitbucket.org/fusemail/fm-lib-commons-golang/sys"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// exit code constants
const (
	SUCCESS = iota
	FAIL
	Port = 9091
)

var tokenString string

var flagErrorCode int

var HttpErrors map[int]string

var options struct {
	System      sys.Options               `group:"Default System Options"`
	Application server.ApplicationOptions `group:"Default Application Server Options"`

	// Plus your own opts. (remove this for command-line app)

}

func init() {

}

func main() {

	flag.IntVar(&flagErrorCode, "erro", 200, "Status code to return")
	HttpErrors = make(map[int]string)
	HttpErrors[http.StatusInternalServerError] = "Some error in the server"
	HttpErrors[http.StatusForbidden] = "User not Authenticated"
	HttpErrors[http.StatusUnauthorized] = "User not Authorized"

	flag.Parse()

	log.Infof("flagErrorCode : %d", flagErrorCode)

	// Set the proper exist code before exit. DO NOT EXIT DIRECTLY.
	exitCode := FAIL
	defer func() { os.Exit(exitCode) }()

	// Override system logging used by base libraries.
	system := sys.NewLogger(log.StandardLogger())

	// Setup system.
	sys.SetLogger(system)
	sys.SetupOptions(&options, &options.System)

	// remove all the code below in this function if you are building a command-line app

	// to display README as service home page
	// bindata.Setup(Asset, AssetDir, AssetNames)

	router := mux.NewRouter()

	router.HandleFunc("/login", HandleLogin)
	router.HandleFunc("/report", HandleReport)

	server.SetLogger(system)
	_, ok := server.Setup(&server.Config{
		Port:    Port,
		UseSSL:  options.Application.SSL,
		SSLCert: options.Application.SSLCert,
		SSLKey:  options.Application.SSLKey,
		Router:  router,
	})
	if !ok {
		log.Error("Server setup returned false")
		return
	}

	httpHandler := httphandler.New(router, middleware.Common(), options.Application.Limit)
	httpHandler.MountDefaultEndpoints(options.Application)

	// Setup metrics.
	metrics.SetLogger(system)
	metrics.Register() // No additional metrics.
	metrics.Serve()

	// Setup health with dependencies.
	health.SetLogger(system)
	health.Register(
		&health.Dependency{
			Name: "Sample",
			Desc: "Sample",
			Item: &deps.Sample{},
		},
	)
	health.Serve()

	// Start serving the application
	server.Serve()

	// Consul get datacenter
	if options.Application.ConsulRegistration {
		serviceConsul := &server.Service{
			Name:             options.Application.ConsulName,
			RegistrationHost: options.Application.ConsulHost,
			Port:             Port,
		}
		serviceConsul.MustRegister()
		log.Debug("registered to consul: ", serviceConsul)
		defer serviceConsul.Deregister() // nolint:errcheck
	}

	log.Info("DEMO sucessfully started.")

	sys.BlockAndFunc(func(os.Signal) {
		server.ShutdownAll() // ShutdownAllWithTimeout, ShutdownAllWithContext.
		// All dependencies that need to be close
	})

	exitCode = SUCCESS
}

func HandleLogin(w http.ResponseWriter, r *http.Request) { //nolint

	token, _ := uuid.NewV4()
	tokenString = token.String()
	rawc := "JSESSIONID=" + tokenString

	expire := time.Now().AddDate(0, 0, 1)
	cookie := http.Cookie{
		Name:       "JSESSIONID",
		Value:      tokenString,
		Path:       "/report",
		Domain:     "localhost:9091",
		Expires:    expire,
		RawExpires: expire.Format(time.UnixDate),
		MaxAge:     86400,
		Secure:     true,
		HttpOnly:   true,
		Raw:        rawc,
		Unparsed:   []string{rawc},
	}

	http.SetCookie(w, &cookie)

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Login succeded")

}

func HandleReport(w http.ResponseWriter, r *http.Request) { //nolint

	if errorMsg, ok := HttpErrors[flagErrorCode]; ok {
		w.WriteHeader(flagErrorCode)
		io.WriteString(w, errorMsg)
		log.Info("Using input error")
	} else {

		fileName, err := CreateFile()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, err.Error())
		}

		http.ServeFile(w, r, "./output/"+fileName)

	}

}
