package sys

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"bitbucket.org/fusemail/fm-lib-commons-golang/utils"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

/*
SetupEnvironment loads environment variables once:
	dev, prod, test, etc.
This "env" type evaluates from following sources in order:
	- Command line option -e or --environment.
	- ENVIRONMENT environment variable.
	- Default as "dev".
Once "env" type is identified, it loads environment variables from the following sources:
	- {local}/{program-name}-{Env}.env
	- {EnvPath}/{program-name}-{Env}.env
Finally the environment is applied to the provided options.
*/
func SetupEnvironment(environment string, environmentPath string, options interface{}) {
	log.WithFields(logrus.Fields{"environment": environment, "environmentPath": environmentPath}).Info("setup environment")

	k := "ENVIRONMENT"
	// At this point, environment already holds the single valid env var, from these sources in order:
	//    command line option, ENVIRONMENT var, "dev" default.
	env := utils.GetEnv(k, environment)

	if env != environment {
		env = environment
		log.WithField("env", env).Info("setting ENVIRONMENT var")
		err := os.Setenv(k, env)

		if err != nil {
			log.WithFields(logrus.Fields{"env": env, "k": k, "err": err}).Panic("failed to set ENVIRONMENT var")
		}
	}

	// Fix (clean) env path if necessary.
	newEnvPath := path.Clean(environmentPath)

	if newEnvPath != environmentPath {
		log.WithFields(logrus.Fields{"environmentPath": environmentPath, "newEnvPath": newEnvPath}).Info("Fixing env path")
		environmentPath = newEnvPath
	}

	// Load env file from local/current directory.
	// Do not use os.Args[0], which returns a temporary folder when "go run" is executed.
	if fpath, err := os.Getwd(); err == nil {
		loadEnvironment(env, fpath)
	} else {
		// Do not fail, just warn.
		log.WithField("err", err).Info("failed to get current directory (so ignore local env file)")
	}

	// Load env file from deployment server @ EnvPath.
	loadEnvironment(env, environmentPath)

	// Apply environment variables.
	parseOptions(options, true)

	// Log final applied options, after latest parsing and environments.
	log.WithField("options", utils.StrPlus(options)).Debug("applied options")
}

func loadEnvironment(env string, fpath string) {
	log.WithFields(logrus.Fields{"env": env, "fpath": fpath}).Debug("@loadEnvironment")

	fname := fmt.Sprintf("%v-%v.env", filepath.Base(os.Args[0]), env)
	src := filepath.Join(fpath, fname)
	log.WithField("src", src).Info("loading env file")

	if err := godotenv.Load(src); err != nil {
		log.WithFields(logrus.Fields{"src": src, "err": err}).Info("environment file at src location not loaded")
	}
}
