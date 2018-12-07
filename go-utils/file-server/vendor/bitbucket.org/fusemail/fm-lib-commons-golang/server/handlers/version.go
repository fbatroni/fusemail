package handlers

import (
	"net/http"

	"bitbucket.org/fusemail/fm-lib-commons-golang/server"
	"bitbucket.org/fusemail/fm-lib-commons-golang/sys"
)

// Version provides version/build data.
func Version() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.WriteJSON(w, sys.BuildInfo)
	})
}
