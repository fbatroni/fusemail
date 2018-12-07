/*
Package client provides web client capabilities.
Duties:
- Provides Config to configure web clients.
- Provides ParsedResponse to parse web client responses.
- Holds default Client in Default var.
- Provides both package/Default & Client levels for HTTP methods:
    Get / Post / PostForm / Head.
- Provides Client functionality, with proper defaults and functionality:
	timeout, log, close, metrics.
- Instruments metrics:
	client_http_request_total, client_http_request_bytes, client_http_request_duration_seconds.
- Generates statistics via Stats.
- Logs pertinent info, including LogRequests option.
*/
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/fusemail/fm-lib-commons-golang/metrics"
	log "github.com/sirupsen/logrus"
)

var (
	// Config holds the web client configuration.
	Config = &struct {
		StatusFailed string `json:"status_failed"`
	}{
		StatusFailed: "X-Failed",
	}

	// Stats contains web client statistics.
	Stats = struct {
		Total   int `json:"total"`
		Current int `json:"curr"`
		Fails   int `json:"fails"`
	}{}

	// Default holds the default web client, with extended functionality and defaults.
	Default = NewClient()
)

// Client extends the base Client struct.
type Client struct {
	*http.Client
	LogRequests bool
}

func (c *Client) String() string {
	return fmt.Sprintf("{%T}", c)
}

// NewClient constructs web client instances, with sensible defaults.
// TBD - consider gorequest instead?: https://github.com/parnurzeal/gorequest
func NewClient() *Client {
	// Reasonable defaults.
	c := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return &Client{
		Client: c,
		// LogRequests: true,
	}
}

// Do handles all method types with headers.
func (c *Client) Do(method, url string, body io.Reader, headers map[string]string) (resp *http.Response, err error) {
	method = strings.ToUpper(method)
	return c.Send(method, url, func() (*http.Response, error) {
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, err
		}
		for key, val := range headers {
			req.Header.Set(key, val)
		}
		return c.Client.Do(req)
	})
}

// Get handles GET calls.
func (c *Client) Get(url string) (*http.Response, error) {
	return c.Send("get", url, func() (*http.Response, error) {
		return c.Client.Get(url)
	})
}

// Post handles POST calls.
func (c *Client) Post(url string, bodyType string, body io.Reader) (*http.Response, error) {
	return c.Send("post", url, func() (*http.Response, error) {
		return c.Client.Post(url, bodyType, body)
	})
}

// PostForm handles POST calls.
func (c *Client) PostForm(url string, data url.Values) (*http.Response, error) {
	return c.Send("post", url, func() (*http.Response, error) {
		return c.Client.PostForm(url, data)
	})
}

// Head handles HEAD calls.
func (c *Client) Head(url string) (resp *http.Response, err error) {
	return c.Send("head", url, func() (*http.Response, error) {
		return c.Client.Head(url)
	})
}

func (c *Client) log(msg string, d *RequestData) {
	if c.LogRequests {
		log.WithField("data", d).Info(msg)
		return
	}
	log.WithField("data", d).Debug(msg)
}

// RequestData holds client request data.
type RequestData struct {
	Request *Request
	URL     string
	Method  string
}

func (d *RequestData) String() string {
	return fmt.Sprintf(
		"{%T: %v %v , %v}",
		d, d.Method, d.URL, d.Request,
	)
}

// Send executes the requested func, with all extra functionality.
// Exported so it can be called indirectly.
func (c *Client) Send(method, url string, fn func() (*http.Response, error)) (*http.Response, error) {
	log.WithFields(log.Fields{"url": url, "method": method}).Debug("@Client.Send")

	d := &RequestData{
		Request: NewRequest(),
		URL:     url,
		Method:  method,
	}

	c.log("sending web client request", d)

	Stats.Total++
	Stats.Current++

	failed := false
	resp, err := fn() // Execute.
	d.Request.Finish()
	status := Config.StatusFailed

	if err == nil {
		// Must NOT log "resp", otherwise error on log debug as json format.
		c.log("received web client response", d)
		// resp.Body.Close() // Do NOT close automatically.
		status = strconv.Itoa(resp.StatusCode)
	} else {
		failed = true
		Stats.Fails++
		c.log("failed web client request: "+err.Error(), d)
	}

	defer func() {
		Stats.Current--

		if !metrics.Served {
			return
		}

		vals := []string{
			url,
			method,
			status,
		}

		metrics.Add(
			"client_http_request_total",
			1,
			vals...,
		)

		if failed {
			return
		}

		metrics.Add(
			"client_http_request_bytes",
			float64(resp.ContentLength),
			vals...,
		)
		metrics.Add(
			"client_http_request_duration_seconds",
			d.Request.Duration.Seconds(),
			vals...,
		)
	}()

	return resp, err
}

// Do redirects to Default.
func Do(method, url string, body io.Reader, headers map[string]string) (resp *http.Response, err error) {
	return Default.Do(method, url, body, headers)
}

// Get redirects to Default.
func Get(url string) (*http.Response, error) {
	return Default.Get(url)
}

// Post redirects to Default.
func Post(url string, bodyType string, body io.Reader) (*http.Response, error) {
	return Default.Post(url, bodyType, body)
}

// PostForm redirects to Default.
func PostForm(url string, data url.Values) (*http.Response, error) {
	return Default.PostForm(url, data)
}

// Head redirects to Default.
func Head(url string) (resp *http.Response, err error) {
	return Default.Head(url)
}

// ParsedResponse holds a parsed response, either JSON or Text.
type ParsedResponse struct {
	Status      int
	Status2xx   bool
	ContentType string
	Bytes       []byte      `json:"-"`
	IsJSON      bool        // In addition to BodyJSON, to support true and error.
	BodyJSON    interface{} `json:"-"`
	BodyText    string      `json:"-"`
}

func (pr *ParsedResponse) String() string {
	typ := "text"
	if pr.IsJSON {
		typ = fmt.Sprintf("json (%T)", pr.BodyJSON)
	}

	return fmt.Sprintf(
		"{%T: %v %v, %v bytes }",
		pr, pr.Status, typ, len(pr.Bytes),
	)
}

// NewParsedResponse constructs ParsedResponse based on provided resp.
// Attempts to unmarshall into &target and assigns to BodyJSON.
// Returns a second part with nil or error.
func NewParsedResponse(resp *http.Response, target interface{}) (*ParsedResponse, error) {
	defer resp.Body.Close()

	pr := &ParsedResponse{
		Status:      resp.StatusCode,
		Status2xx:   resp.StatusCode >= http.StatusOK && resp.StatusCode < 300,
		ContentType: resp.Header.Get("Content-Type"),
	}

	b, err := ioutil.ReadAll(resp.Body)
	pr.Bytes = b

	if err != nil {
		return pr, fmt.Errorf("could not read response body: " + err.Error())
	}

	// Compare against lower, in order to support case-insensitive.
	ctype := strings.ToLower(pr.ContentType)
	pr.IsJSON = strings.HasPrefix(ctype, "application/json")

	if pr.IsJSON {
		err := json.Unmarshal(b, &target)
		pr.BodyJSON = target

		if err != nil {
			return pr, fmt.Errorf("could not parse response as JSON: " + err.Error())
		}
	} else {
		pr.BodyText = string(b)
	}

	return pr, nil
}
