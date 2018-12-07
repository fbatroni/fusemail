package client

import (
	"fmt"
	"time"

	"bitbucket.org/fusemail/fm-lib-commons-golang/utils"
)

// Request represents a web client request.
type Request struct {
	UID      string // Unique identifier of the request, for accurate identification in log files.
	Started  time.Time
	Finished time.Time
	Duration time.Duration `json:"duration"`
}

// NewRequest constructs web client requests, and marks the beginning the request.
func NewRequest() *Request {
	r := &Request{
		Started: time.Now(),
	}

	// UID = elapsed seconds since unix epoch + pointer address.
	r.UID = fmt.Sprintf("%v-%p", r.Started.Unix(), r)

	return r
}

func (r *Request) String() string {
	msStr := "~" // Representation for "in progress".
	if !r.Finished.IsZero() {
		msStr = fmt.Sprintf("%v ms", utils.ElapsedMillis(r.Started, r.Finished))
	}
	return fmt.Sprintf(
		"{%T: (UID:%v) %v since %v}",
		r, r.UID, msStr, r.Started,
	)
}

// Finish marks the end of the request.
func (r *Request) Finish() {
	r.Finished = time.Now()
	r.Duration = time.Since(r.Started)
}
