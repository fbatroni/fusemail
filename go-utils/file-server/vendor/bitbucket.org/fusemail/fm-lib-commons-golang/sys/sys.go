/*
Package sys provides sys/global capabilities across all packages.
Duties:
- Provides Options to embed for system/environment options.
- Setup system/environment via SetupOptions, as follows:
	1. Apply environment options (Env & EnvPath) from command line or environment or defaults.
	2. Search and apply environment files based on Env & EnvPath values:
		{local}/{program}-{Env}.env
		{EnvPath}/{program}-{Env}.env
	3. Apply all other options (except Env & EnvPath) from command line or environment or defaults.
- Prints HELP and exits, via -h | --help flag, or on any flag error.
- Prints VERSION and exits, via -v | --version flag.
- Processes all sys (environment and global) options (see HELP output):
	Env, EnvPath, Version, Debug.
- Holds global data, including BuildInfo version vars, to be provided during the build process.
	- Usage in Makefile:
		SYSPKG=bitbucket.org/fusemail/fm-lib-commons-golang/sys
		go build -o build/fm-service-DEMO -ldflags "\
			-X $(SYSPKG).Version=$(VERSION) \
			-X $(SYSPKG).GitHash=$(GITHASH) \
			-X $(SYSPKG).BuildStamp=$(BUILDDATE) \
		"
	- Usage with go run (just for testing purposes):
	  Must replace SYSPKG with the full path from GOPATH (including /vendor):
	  		bitbucket.org/fusemail/fm-service-DEMO/vendor/bitbucket.org/fusemail/fm-lib-commons-golang/sys
		go run -ldflags "
			-X SYSPKG.Version=vVersion
			-X SYSPKG.GitHash=vGitHash
			-X SYSPKG.BuildStamp=vBuildStamp
		" main.go -v
- Configures logging according to sys opts, including formatters and debug hooks.
- Logs exit data since Serve (LogExit), and exits with the provided code (Exit).
- Provides BlockAndExit to automatically block/wait until signals and then exit with code.
- Provides BlockAndFunc for more granular control.
- Avoids the need to import extra packages in your app, e.g. go-flags, godotenv.
- Logs pertinent info.
*/
package sys

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

// Started timestamp of the system
var Started time.Time

func init() {
	ts := "2006-01-02T15:04:05.999999Z07:00"

	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: ts,
			DisableColors:   true})
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: ts})
	}

	Started = time.Now()

	BuildInfo["version"] = Version
	BuildInfo["git_hash"] = GitHash
	BuildInfo["build_stamp"] = BuildStamp

	built = BuildStamp != unsetBuild
}

const unsetBuild = "UNSET"

// Version variables set during build process in Makefile.
// See example in package comment.
// Must be direct variables, not within a struct.
var (
	Version    = "DO NOT USE IN PRODUCTION"
	GitHash    = unsetBuild
	BuildStamp = unsetBuild
)

var (
	// BuildInfo holds the version/build info.
	BuildInfo = make(map[string]string)

	// built indicates whether sys has been built.
	// Unexported, set only from sys init, and read only via Built func.
	built = false

	// Package logger, set with SetLogger.
	log = logrus.StandardLogger()
)

// Built returns whether sys has been built.
func Built() bool {
	return built
}
