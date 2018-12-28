package deps

// Provides HealthAPI struct to check health against any fusemail health endpoint.

import (
	"errors"
	"fmt"

	"bitbucket.org/fusemail/fm-lib-commons-golang/client"
	"bitbucket.org/fusemail/fm-lib-commons-golang/health"
	log "github.com/sirupsen/logrus"
)

// MaxResponseSize refers to maximum response size.
// If bigger than this size, then this API truncates response and forces text (even if JSON data).
const MaxResponseSize = 30 * 1000

// HealthAPI provides health struct.
type HealthAPI struct {
	URL string `json:"url"`
}

// Check checks health.
// Returns map (optional config/state) and error (nil if healthy).
func (d *HealthAPI) Check() (map[string]interface{}, error) {
	log.WithField("d", d).Debug("@Check")

	state := make(map[string]interface{}) // No extra data by default.

	resp, err := client.Get(d.URL)
	if err != nil {
		return state, err
	}

	parsedResp, err := client.NewParsedResponse(resp, nil)
	log.WithFields(log.Fields{"parsedResp": parsedResp, "err": err}).Debug("Parsed response")

	state["status"] = parsedResp.Status
	state["content_type"] = parsedResp.ContentType

	size := len(parsedResp.Bytes)
	state["response_size"] = size

	if err != nil {
		return state, err
	}

	// NOTE that "response" could grow exponentially, e.g. when services check health against each other.
	// So include it conditionally depending on current response size, otherwise truncate and force text.
	if size <= MaxResponseSize && parsedResp.IsJSON {
		state["response"] = parsedResp.BodyJSON
	} else {
		txt := string(parsedResp.Bytes) // Force text, even if JSON.
		if size > MaxResponseSize {
			txt = txt[:MaxResponseSize]
		}
		state["response"] = txt
	}

	if parsedResp.Status != health.StatusHealthy {
		emsg := fmt.Sprintf("Unhealthy HealthAPI with status: %v", parsedResp.Status)

		// Error log level already @ health.
		log.WithFields(log.Fields{"d": d, "size": len(parsedResp.Bytes)}).Debug(emsg)

		return state, errors.New(emsg)
	}

	return state, nil
}
