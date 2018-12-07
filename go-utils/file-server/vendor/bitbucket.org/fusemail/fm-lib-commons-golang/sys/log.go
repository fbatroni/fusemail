package sys

// Provides log setup with hooks.

import (
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var detectEnvironmentFields = []string{"NOMAD_ALLOC_ID", "NOMAD_TASK_NAME", "NOMAD_DC"}

// SetupLogging sets up log context in debug mode.
func SetupLogging(debug bool) {
	log.WithField("debug", debug).Info("setup logging")

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.AddHook(LogContextHook{})
	}
}

// SetLogger sets the system logger used by base libraries
// Use this to stop the system logging for command line utilities or if you just
// wish to support the system level logging
func SetLogger(logger *logrus.Logger) {
	log = logger
}

// NewLogger duplicates the passed in logger settings
func NewLogger(logger *logrus.Logger) *logrus.Logger {
	newLogger := logrus.New()
	if logger != nil {
		newLogger.Out = logger.Out
		newLogger.Formatter = logger.Formatter
		newLogger.Hooks = logger.Hooks
		newLogger.Level = logger.Level
	}
	env := NewLogEnvironmentHook(detectEnvironmentFields...)
	// if we have found one of the environment variables that show we are under nomad
	if len(env.Fields) > 1 {
		newLogger.Hooks.Add(env)
	}
	return newLogger
}

// LogEnvironmentHook adds environment fields to every log entry, for extended logging
// under Nomad services
type LogEnvironmentHook struct {
	Fields map[string]string
}

// NewLogEnvironmentHook sets up a new environment logger hook that takes fields
// from the environment and logs them on each line
func NewLogEnvironmentHook(fields ...string) *LogEnvironmentHook {
	l := &LogEnvironmentHook{}
	l.Fields = make(map[string]string)
	for _, field := range fields {
		value := os.Getenv(field)
		if len(value) > 0 {
			l.Fields[field] = value
		}
	}
	hostname, err := os.Hostname()
	if err == nil {
		l.Fields["hostname"] = hostname
	}
	return l
}

// Levels returns permitted log levels
func (l *LogEnvironmentHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.ErrorLevel}
}

// Fire adds all fields configured to the log entry
func (l *LogEnvironmentHook) Fire(entry *logrus.Entry) error {
	for field, value := range l.Fields {
		entry.Data[field] = value
	}
	return nil
}

// LogContextHook provides log levels and format to extend logrus functionality.
// It prepends file name, func name, and line number to log lines,
// specifically for log levels: Debug, Info, Error.
// @ https://github.com/sirupsen/logrus/issues/63#issuecomment-236052137
type LogContextHook struct{}

// Levels returns permitted log levels: Debug, Info and Error.
func (hook LogContextHook) Levels() []logrus.Level {
	// fmt.Println("@LogContextHook.Levels")
	return []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.ErrorLevel}
}

// Fire prepends file name, func name, and line number to log lines.
func (hook LogContextHook) Fire(entry *logrus.Entry) error {
	pc := make([]uintptr, 3, 3)
	cnt := runtime.Callers(6, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()

		if !strings.Contains(name, "/logrus.") {
			file, line := fu.FileLine(pc[i] - 1)
			entry.Data["file"] = path.Base(file)
			entry.Data["func"] = path.Base(name)
			entry.Data["line"] = line
			break
		}
	}

	return nil
}
