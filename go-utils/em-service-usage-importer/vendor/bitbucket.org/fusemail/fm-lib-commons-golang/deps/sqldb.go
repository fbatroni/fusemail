package deps

// Provides SQLDB struct to check health against SQL database.

import (
	"database/sql"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/go-sql-driver/mysql"
)

// SQLDB provides health struct.
type SQLDB struct {
	Conf *mysql.Config `json:"-"` // Hide Passwd.

	// Pool size: maximum number of open connections to the database.
	// If <= 0, then unlimited.
	// Default is 0 (unlimited).
	MaxOpenConns int

	// Maximum number of connections in the idle connection pool.
	// If <= 0, no idle connections are retained.
	MaxIdleConns int

	DB *sql.DB `json:"-"` // Live database handle/connection.

	// Reserved fields, do NOT set directly.
	DSN       string `json:"-"` // Private - data source name, showing masked instead.
	MaskedDSN string `json:"masked_dsn"`

	traceFunc []QueryTraceFunc
}

// NewSQLDB constructs SQL database instances.
func NewSQLDB(conf *mysql.Config) *SQLDB {
	d := &SQLDB{Conf: conf}

	// Prepare main DSN.
	d.DSN = conf.FormatDSN()

	// Prepare masked DSN (without password) for state/logging purposes.
	tmpPasswd := conf.Passwd
	conf.Passwd = "***"
	d.MaskedDSN = conf.FormatDSN()
	conf.Passwd = tmpPasswd

	log.WithFields(log.Fields{"dsn": d.MaskedDSN}).Debug("@NewSQLDB")
	return d
}

// Check checks health.
// Returns map (optional config/state) and error (nil if healthy).
func (d *SQLDB) Check() (map[string]interface{}, error) {
	log.WithField("d", d).Debug("@Check")

	// Initialize state with optional extra data.
	state := make(map[string]interface{})
	// state["dsn"] = d.MaskedDSN

	err := d.Connect()

	if err != nil {
		return state, err
	}

	// Update state via queries.
	queries := []string{
		"SHOW VARIABLES LIKE 'report_host'",
		"SHOW VARIABLES LIKE 'version'",
		"SHOW STATUS LIKE 'threads_connected'",
	}

	for _, sql := range queries {
		var name, value string

		row := d.DB.QueryRow(sql)

		if row == nil {
			return state, errors.New("failed to query row")
		}

		err := row.Scan(&name, &value)

		if err != nil {
			return state, err
		}

		state["sql: "+name] = value
	}

	return state, nil // Healthy.
}

// Connect connects to the database, or does nothing if already connected.
// Return error or nil.
func (d *SQLDB) Connect() error {
	connected := d.DB != nil
	log.WithFields(log.Fields{"connected": connected}).Debug("@Connect")

	if connected {
		return nil
	}

	db, err := sql.Open("mysql", d.DSN)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	if d.MaxOpenConns != 0 {
		db.SetMaxOpenConns(d.MaxOpenConns)
	}

	if d.MaxIdleConns != 0 {
		db.SetMaxIdleConns(d.MaxIdleConns)
	}

	// Store it only if no errors after Ping.
	d.DB = db

	return nil
}

// Trace applies list of trace function on QueryTrace details.
func (d *SQLDB) Trace(action string, start time.Time, err error) {
	t := &QueryTrace{
		Action: action,
		Start:  start,
		Error:  err,
	}
	for _, f := range d.traceFunc {
		f(t)
	}
}

// AddTracer appends new query trace function to the existing list.
func (d *SQLDB) AddTracer(tracer QueryTraceFunc) {
	d.traceFunc = append(d.traceFunc, tracer)
}
