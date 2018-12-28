package utils

// Provides environment utilities.

import (
	"os"
	"strings"
)

// GetEnv returns value of the environment variable named "key"; returns "def" if not set.
func GetEnv(k, def string) string {
	v, found := os.LookupEnv(k)
	if !found {
		v = def
	}
	return v
}

// GetEnvList returns a list of values from an environment variable, separated by commas; returns def if not set
func GetEnvList(k string, separator string, def []string) []string {
	var values []string
	v, found := os.LookupEnv(k)
	if !found {
		values = def
	} else {
		values = strings.Split(v, separator)
	}
	return values
}
