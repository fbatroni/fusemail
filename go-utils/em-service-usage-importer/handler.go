package main

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"bitbucket.org/fusemail/em-lib-common-usage/common"
	"bitbucket.org/fusemail/em-lib-common-usage/importer"
	"bitbucket.org/fusemail/fm-lib-commons-golang/httphandler"
	"bitbucket.org/fusemail/fm-lib-commons-golang/server"
	"bitbucket.org/fusemail/fm-lib-commons-golang/server/middleware"

	log "github.com/sirupsen/logrus"
)

var processing bool

func StartImportStep(httpHandler *httphandler.HTTPHandler, importerStep importer.ImporterStep, vendorMapper importer.VendorMapper) http.Handler { // nolint
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		t := &httphandler.RequestTrace{
			Logger: middleware.GetContextLogger(r.Context()),
			Action: "StartImporterStep",
		}
		defer func() {
			if rc := recover(); rc != nil {
				t.Error = fmt.Errorf("%s : %s", rc, debug.Stack())
			}

			httpHandler.Trace(t)

			if t.Error != nil {
				t.Logger.Error(t.Error.Error())
			}
		}()

		if processing {
			server.WriteJSONWithStatus(w, "Serve is busy. Try again later", http.StatusServiceUnavailable)
		} else {

			processing = true

			//Execute as routine. The request gets no response about the process result only response for
			//whether the request was accepted or not
			go func() {

				defer func() {
					processing = false
				}()

				err := importerStep.ExecuteImportStep(vendorMapper)

				if err != nil {
					if err == common.ErrNoFileToTransform {
						log.Info(err.Error())
					} else {
						t.Error = err
					}
				}

			}()

			server.WriteJSONWithStatus(w, "Import Step request enqueued", http.StatusOK)

		}

	})
}
