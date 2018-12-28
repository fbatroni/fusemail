package handlers

import (
	"fmt"
	"net/http"
	"os"

	"bitbucket.org/fusemail/fm-lib-commons-golang/server/middleware"
)

const (
	ContentTypeHTML = "text/html"
)

// AssetFunc represents the func to load files, usually "bindata.LoadFile".
type AssetFunc func(string) ([]byte, error)

func handleDocsText(w http.ResponseWriter, r *http.Request, fn AssetFunc, filename, header, footer, contentType string) {
	logger := middleware.GetContextLogger(r.Context())
	logger.WithField("filename", filename).Debug("@docsHTMLHandler")

	w.Header().Add("Content-Type", contentType)

	data, err := fn(filename)

	if err != nil {
		logger.Error(err)
		http.Error(w, fmt.Sprintf("%s: asset not found", filename), http.StatusNotFound)
		return
	}

	w.Write([]byte(header))
	w.Write(data)
	w.Write([]byte(footer))
}

// DocsMarkdown handles markdown doc.
func DocsMarkdown(fn AssetFunc, filename string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := fmt.Sprintf(
			`<!DOCTYPE html><html><title>%s</title><xmp theme="united" style="display:block;">`,
			os.Args[0])

		footer := `</xmp><script src="http://strapdownjs.com/v/0.2/strapdown.js"></script></html>`

		handleDocsText(w, r, fn, filename, header, footer, ContentTypeHTML)
	})
}

// DocsHTML handles HTML doc.
func DocsHTML(fn AssetFunc, filename string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDocsText(w, r, fn, filename, "", "", ContentTypeHTML)
	})
}

// Handle all text-type endpoints
func DocsText(fn AssetFunc, filename, contentType string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDocsText(w, r, fn, filename, "", "", contentType)
	})
}

// Handle redirects
func DocsRedirect(url string, code int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, code)
	})
}
