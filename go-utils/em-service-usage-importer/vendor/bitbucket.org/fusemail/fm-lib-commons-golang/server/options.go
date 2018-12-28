package server

// ApplicationOptions provides default application options to embed.
type ApplicationOptions struct {
	Route string `long:"route" env:"ROUTE" default:"/app" description:"application route"`
	Limit int    `long:"limit" env:"LIMIT" default:"500" description:"application limit for maximum number of concurrent requests; zero to disable"`

	SSL     bool   `long:"ssl" env:"SSL" description:"enable application ssl"`
	SSLCert string `long:"ssl-cert" env:"SSL_CERT" description:"application ssl certificate path"`
	SSLKey  string `long:"ssl-key" env:"SSL_KEY" description:"application ssl key path"`

	HealthPort  int    `long:"health-port" env:"HEALTH_PORT" description:"port number for health and metrics"`
	HealthRoute string `long:"health-route" env:"HEALTH_ROUTE" default:"/health" description:"route for health"`

	MetricsRoute string `long:"metrics-route" env:"METRICS_ROUTE" default:"/metrics" description:"route for metrics"`

	VersionRoute string `long:"version-route" env:"VERSION_ROUTE" default:"/version" description:"route for version"`

	SysRoute string `long:"sys-route" env:"SYS_ROUTE" default:"/sys" description:"route for sys"`

	DocRoute string `long:"doc-route" env:"DOC_ROUTE" default:"/" description:"route for documentation"`
	DocFile  string `long:"doc-file" env:"DOC_FILE" default:"README.md" description:"file name for documentation"`

	APIRoute string `long:"api-route" env:"API_ROUTE" default:"/api" description:"route for API"`
	APIFile  string `long:"api-file" env:"API_FILE" default:"docs/dist/api.html" description:"file name for API"`

	TraceRoute string `long:"trace-route" env:"TRACE_ROUTE" default:"/trace" description:"route for trace"`

	LogRoute string `long:"log-route" env:"LOG_ROUTE" default:"/log" description:"route for log"`
	LogCount int    `long:"log-count" env:"LOG_COUNT" default:"500" description:"number of log lines to keep in buffer"`

	ConsulRegistration bool     `long:"consul" env:"CONSUL" description:"register service with consul"`
	ConsulHost         string   `long:"consul-host" env:"CONSUL_HOST" description:"consul host:port to use" default:"localhost:8500"`
	ConsulName         string   `long:"consul-name" env:"CONSUL_NAME" description:"service name registered in consul, default to program name"`
	ConsulTags         []string `long:"consul-tag" env:"CONSUL_TAGS" env-delim:"," description:"consul registration custom tags" default:"prometheus_exporter"`
}
