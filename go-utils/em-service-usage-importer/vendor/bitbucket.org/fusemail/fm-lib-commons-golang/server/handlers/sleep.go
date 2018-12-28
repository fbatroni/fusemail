package handlers

import (
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/fusemail/fm-lib-commons-golang/server"
	"bitbucket.org/fusemail/fm-lib-commons-golang/server/middleware"
	"bitbucket.org/fusemail/fm-lib-commons-golang/utils"
)

// Sleep sleeps for (?secs=X) or for the default duration.
func Sleep(sleepFor time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := middleware.GetContextLogger(r.Context())

		if secs, err := strconv.Atoi(r.URL.Query().Get("secs")); err == nil {
			sleepFor = time.Duration(secs) * time.Second
		}

		logger.Info("sleeping")
		time.Sleep(sleepFor)
		logger.WithField("secs", sleepFor).Info("slept")

		server.WriteJSON(w, utils.OpenMap{
			"route":    r.URL.Path,
			"duration": sleepFor.Seconds(),
		})
	})
}
