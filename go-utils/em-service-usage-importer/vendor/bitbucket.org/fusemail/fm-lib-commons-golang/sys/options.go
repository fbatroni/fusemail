package sys

// Provides options setup with env.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"bitbucket.org/fusemail/fm-lib-commons-golang/utils"
	"github.com/sirupsen/logrus"
	"github.com/jessevdk/go-flags"
)

/*
MaskedString is a protected string, in order to not leak private data in logs and json:
	- masked on logging "applied options" (by implementing Stringer).
	- masked on handling /sys endpoint (by overriding MarshalJSON).
Protects all private options, e.g. passwords.
Must be converted to primitive string when using it:
	string(maskedOption).
*/
type MaskedString string

func (s MaskedString) String() string {
	return "***"
}

// MarshalJSON supports json log format.
func (s MaskedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// Options provides default environment and system options to embed.
type Options struct {
	// Environment options.
	Env     string `short:"e" long:"environment" env:"ENVIRONMENT" default:"dev" description:"environment (prod or dev)"`
	EnvPath string `long:"environment-path" env:"ENVIRONMENT_PATH" description:"environment path, default to /etc/fusemail/{binary name}"`

	// System options.
	Version bool `short:"v" long:"version" description:"output version information"`
	Debug   bool `short:"d" long:"debug" env:"DEBUG" description:"enable debugging"`
}

// SetupOptions sets up options from command line and environment.
// It requires the full options and the specific system options as parameters.
func SetupOptions(allOptions interface{}, sysOptions *Options) {
	log.WithField(
		"start_time", Started).Printf("starting up: %v",
		Started.Format("15:04:05 MST 2006/01/02"))
	log.WithFields(logrus.Fields{
		"allOptions": allOptions,
		"sysOptions": sysOptions}).Debug("@SetupOptions")

	PrepareOptions(allOptions)
	CommandOptionVersion(sysOptions.Version)
	SetupLogging(sysOptions.Debug)
	if sysOptions.EnvPath == "" {
		sysOptions.EnvPath = fmt.Sprintf("/etc/fusemail/%s", filepath.Base(os.Args[0]))
	}
	SetupEnvironment(sysOptions.Env, sysOptions.EnvPath, allOptions)
}

// PrepareOptions applies HELP and validates options.
func PrepareOptions(options interface{}) {
	log.WithField("options", options).Debug("@PrepareOptions")
	parseOptions(options, false)
}

// parseOptions parses options and handles errors according to envReady.
// Exits on any unhandled error.
func parseOptions(options interface{}, envReady bool) {
	log.WithFields(logrus.Fields{"options": options, "envReady": envReady}).Debug("@parseOptions")

	flagOpts := flags.HelpFlag
	if envReady {
		flagOpts = flags.Default
	}

	parser := flags.NewParser(options, flags.Options(flagOpts))
	_, err := parser.Parse()

	// If clean, no further checks and validations.
	if err == nil {
		return
	}

	writeHelpAndExit := func(exitCode int) {
		log.WithField("exitCode", exitCode).Debug("@writeHelpAndExit")
		parser.WriteHelp(os.Stdout)
		Exit(exitCode)
	}

	if flagsErr, ok := err.(*flags.Error); ok {
		// If environment is not yet ready, then ignore "required" error.
		if flagsErr.Type == flags.ErrRequired && !envReady {
			return
		}
		// Handle help flag (-h | --help), print HELP for ALL options (including embeds) and exit.
		if flagsErr.Type == flags.ErrHelp {
			log.Info("Print HELP (-h | --help) and exit")
			writeHelpAndExit(0)
		}
	}

	log.WithFields(logrus.Fields{"options": utils.StrPlus(options), "err": err}).Error("error while parsing options")
	writeHelpAndExit(1)
}

// CommandOptionVersion logs version information, and optionally exits.
func CommandOptionVersion(exit bool) {
	// If exit, then attempt to output the pure version info in JSON format
	// as a new line, nothing else (including timestamp).
	// Falls back to full logging in case of any error.
	if exit {
		byts, err := json.Marshal(BuildInfo)
		if err == nil {
			_, err = fmt.Fprintln(os.Stdout, string(byts))
			if err == nil {
				// Low level exit, not sys'.
				os.Exit(0)
			}
		}
	}

	log.WithField("build_info", BuildInfo).Info("version")

	// Fallback case (after above full logging) on exit.
	if exit {
		Exit(0)
	}
}
