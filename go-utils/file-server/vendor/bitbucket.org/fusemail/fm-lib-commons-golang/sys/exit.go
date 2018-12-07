package sys

// Provides exit capabilities.

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"bitbucket.org/fusemail/fm-lib-commons-golang/utils"
	"github.com/sirupsen/logrus"
)

// Exit logs and exits the app with the provided exit code.
func Exit(code int) {
	now := time.Now()
	var ms int64
	if !Started.IsZero() {
		ms = utils.ElapsedMillis(Started, now)
	}
	log.WithFields(logrus.Fields{
		"started":     Started,
		"finished":    now,
		"duration_ms": ms,
		"exit_code":   code,
	}).Info("exiting")
	os.Exit(code)
}

/*
BlockAndExit blocks/waits forever (once all web servers/listeners have been served),
and then immediately exits with the provided code.
Does NOT shutdown servers gracefully.
Use BlockAndFunc to shutdown servers gracefully via server.ShutdownAll*.
*/
func BlockAndExit(code int, signals ...os.Signal) {
	log.WithFields(logrus.Fields{"code": code, "signals": signals}).Debug("@BlockAndExit")
	BlockAndFunc(func(s os.Signal) {
		Exit(code)
	}, signals...)
}

// BlockAndFunc blocks/waits forever (once all web servers/listeners have been served),
// and then executes the provided func.
func BlockAndFunc(fn func(os.Signal), signals ...os.Signal) {
	log.WithField("signals", signals).Debug("@BlockAndFunc")

	c := make(chan os.Signal)
	if len(signals) == 0 {
		signals = append(signals, os.Interrupt, syscall.SIGTERM)
	}
	log.WithFields(logrus.Fields{"signals": signals}).Info("blocking")
	signal.Notify(c, signals...)

	// Wait for signal.
	s := <-c
	log.WithFields(logrus.Fields{"signal": s}).Info("signal received")
	fn(s)
}
