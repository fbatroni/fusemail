// Package main provides the foundation for your service.
package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"bitbucket.org/fusemail/em-lib-common-usage/common"
	"bitbucket.org/fusemail/em-lib-common-usage/db"
	"bitbucket.org/fusemail/em-lib-common-usage/db/sqlxdb"
	"bitbucket.org/fusemail/em-lib-common-usage/helper/metricvec"
	"bitbucket.org/fusemail/em-lib-common-usage/importer"
	"bitbucket.org/fusemail/fm-lib-commons-golang/bindata"
	"bitbucket.org/fusemail/fm-lib-commons-golang/health"
	"bitbucket.org/fusemail/fm-lib-commons-golang/httphandler"
	"bitbucket.org/fusemail/fm-lib-commons-golang/metrics"
	"bitbucket.org/fusemail/fm-lib-commons-golang/server"
	"bitbucket.org/fusemail/fm-lib-commons-golang/server/middleware"
	"bitbucket.org/fusemail/fm-lib-commons-golang/sys"
	"bitbucket.org/fusemail/fm-lib-commons-golang/trace"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// exit code constants
const (
	SUCCESS = iota
	FAIL
)

var options struct {
	System      sys.Options               `group:"Default System Options"`
	Application server.ApplicationOptions `group:"Default Application Server Options"`

	// Plus your own opts. (remove this for command-line app)
	Port        int    `long:"port" env:"PORT" description:"application port" default:"8080"`
	InputFolder string `long:"input-folder" env:"INPUT_FOLDER" description:"folder where the importer fetches files" default:"transforms"`
	SourceID    int64  `long:"source-id" env:"SOURCE_ID" description:"Source for usage data" default:"0"`
	StepTypeID  int64  `long:"step-type-id" env:"STEPTYPE_ID" description:"The id for this transform type" default:"0"`
	User        string `long:"user" env:"USER" description:"User that is executing this download step" default:""`

	//Billing database config
	BillingDBUser string `long:"billing database username" env:"DATABASE_USER" description:"billing database user name" default:""`
	BillingDBPass string `long:"billing database password" env:"DATABASE_PASS" description:"billing database user password" default:""`
	BillingDBHost string `long:"billing database host" env:"DATABASE_HOST" description:"billing database host url" default:"127.0.0.1:3306"`
	BillingDBName string `long:"billing database name" env:"DATABASE_NAME" description:"billing database name" default:""`
}

func main() {
	// Set the proper exist code before exit. DO NOT EXIT DIRECTLY.
	start := time.Now()
	exitCode := FAIL
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("%s : %s", r, debug.Stack())
		}
		log.WithField("duration", time.Since(start).String()).Info("program completed")
		os.Exit(exitCode)
	}()

	// Override system logging used by base libraries.
	system := sys.NewLogger(log.StandardLogger())

	// Setup system.
	sys.SetLogger(system)
	sys.SetupOptions(&options, &options.System)

	// remove all the code below in this function if you are building a command-line app

	// to display README as service home page
	bindata.Setup(Asset, AssetDir, AssetNames)

	router := mux.NewRouter()

	server.SetLogger(system)
	_, ok := server.Setup(&server.Config{
		Port:    options.Port,
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
	httpHandler.MountDefaultEndpoints(options.Application, options.System)
	httpHandler.AddTracer(httphandler.LogError())
	httpHandler.AddTracer(httphandler.ReportErrorMetric(metrics.ErrorCounter))

	// Setup metrics.
	metrics.SetLogger(system)
	setupAndServeMetrics()

	//Setting up Billing DB
	billingDB, err := setupBillingDB()
	if err != nil {
		log.Fatal(err)
		return
	}

	//Seting up ImportStep Service
	importStep, err := setupImporterStep(billingDB)
	if err != nil {
		log.Fatal(err)
		return
	}

	vendorMapper, err := setupVendorMapper()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Setup health with dependencies.
	health.SetLogger(system)
	health.Register(
		&health.Dependency{
			Name: "BillingDB",
			Desc: "Billing Database Interface",
			Item: billingDB,
		},
	)
	health.Serve()

	// Start serving the application
	server.Serve()

	limiter := httpHandler.Limiter()
	router.Handle("/start-job", limiter.With(negroni.Wrap(StartImportStep(httpHandler, importStep, vendorMapper)))).Methods(http.MethodGet)

	// Consul get datacenter
	if options.Application.ConsulRegistration {
		serviceConsul := &server.Service{
			Name:             options.Application.ConsulName,
			RegistrationHost: options.Application.ConsulHost,
			Port:             options.Port,
		}
		serviceConsul.MustRegister()
		log.Debug("registered to consul: ", serviceConsul)
		defer serviceConsul.Deregister() // nolint:errcheck
	}

	log.Info("Importer sucessfully started.")

	sys.BlockAndFunc(func(os.Signal) {
		server.ShutdownAll() // ShutdownAllWithTimeout, ShutdownAllWithContext.
		// All dependencies that need to be close
	})

	exitCode = SUCCESS
}

func setupVendorMapper() (importer.VendorMapper, error) {

	var vendorMapper importer.VendorMapper

	yamlFile, err := ioutil.ReadFile("config/vendor-mapper.yml")
	if err != nil {
		return vendorMapper, err
	}

	err = yaml.Unmarshal(yamlFile, &vendorMapper)
	if err != nil {
		return vendorMapper, err
	}

	return vendorMapper, nil

}

func setupBillingDB() (*db.BillingDB, error) {

	dbOptions := sqlxdb.Options{
		DBUser: options.BillingDBUser,
		DBPass: options.BillingDBPass,
		DBAddr: options.BillingDBHost,
		DBName: options.BillingDBName,
	}

	sqlxDB := sqlxdb.New(dbOptions)

	err := sqlxDB.Connect()

	return sqlxDB, err

}

func setupImporterStep(billingDB *db.BillingDB) (importer.ImporterStep, error) {

	//Creates the options for the DownloadStep
	commonOptions := common.Options{
		SourceID:    options.SourceID,
		StepTypeID:  options.StepTypeID,
		InputFolder: options.InputFolder,
		User:        options.User,
	}

	stepService := common.NewStepService(billingDB, commonOptions)

	importerStep, err := importer.New(stepService, commonOptions, billingDB)
	importerStep.AddTracers(trace.Logger(), trace.ErrorMetric(metrics.ErrorCounter))

	return importerStep, err
}

func setupAndServeMetrics() {

	vectors := metrics.NewMetricVectors([]*metrics.Metric{
		metrics.ErrorCounter,
		metricvec.MetricActionDurationTimer,
		metricvec.MetricDownloadTimer,
		metricvec.MetricSQLQueryCounter,
		metricvec.MetricSQLQueryTimer,
	})

	metrics.Register(vectors...)
	metrics.Serve()

	metrics.ErrorCounter.Add(0, "init counter", "not an error")
}
