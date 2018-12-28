package handlers

import (
	"net"
	"net/http"

	"bitbucket.org/fusemail/fm-lib-commons-golang/client"
	"bitbucket.org/fusemail/fm-lib-commons-golang/health"
	"bitbucket.org/fusemail/fm-lib-commons-golang/metrics"
	"bitbucket.org/fusemail/fm-lib-commons-golang/server"
	"bitbucket.org/fusemail/fm-lib-commons-golang/sys"
	"bitbucket.org/fusemail/fm-lib-commons-golang/utils"
)

// Sys provides sys & stats data.
func Sys(options interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO - better format for Duration values on json marshal, instead of the default nanoseconds.
		// Either wrap Duration and implement MarshalJSON method to satisfy Marshaler interface.
		// Or use the "alias" approach @ http://choly.ca/post/go-json-marshalling/
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			ip = r.RemoteAddr
		}

		server.WriteJSON(w, utils.OpenMap{
			"version": sys.BuildInfo,

			"options": options,

			"sys": utils.OpenMap{},

			"client_ip": ip,

			"server": utils.OpenMap{
				"servers": server.Servers,
				"stats":   server.Stats,
			},

			"client": utils.OpenMap{
				"config": client.Config,
				"stats":  client.Stats,
			},

			"health": utils.OpenMap{
				"served": health.Served,
				"config": health.Config,
				"stats":  health.Stats,
			},

			"metrics": utils.OpenMap{
				"served": metrics.Served,
				"stats":  metrics.Stats,
			},
		})
	})
}
