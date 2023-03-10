/*
   Pilot Control Service
   Copyright (C) 2022-Present SouthWinds Tech Ltd - www.southwinds.io

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"net/http"
	h "southwinds.dev/http"
	"southwinds.dev/pilotctl/core"
)

func main() {
	godotenv.Load(".env")
	// creates a generic http server
	s := h.New("pilotctl", core.Version)
	// add handlers
	s.Http = func(router *mux.Router) {
		// enable encoded path  vars
		router.UseEncodedPath()
		// middleware
		router.Use(s.LoggingMiddleware)
		router.Use(s.AuthenticationMiddleware)
		router.Use(mux.CORSMethodMiddleware(router))

		// we have to process cfg here and not pass it to
		// CorsMiddlewhare because it will create
		// circular dependency for now
		cfg := core.NewConf()
		origin := cfg.GetCorsOrigin()
		headers := cfg.GetCorsHeaders()
		router.Use(s.CorsMiddleware(origin, headers))

		// pilot http handlers
		router.HandleFunc("/ping", pingHandler).Methods(http.MethodPost)
		router.HandleFunc("/register", registerHandler).Methods(http.MethodPost)
		router.HandleFunc("/cve/upload", cveReportExportHandler).Methods(http.MethodPost)
		router.HandleFunc("/metrics/{channel}", metricsHandler).Methods(http.MethodPost)
		router.HandleFunc("/logs/{channel}", logsHandler).Methods(http.MethodPost)

		// apply authorisation to admin user http handlers
		router.Handle("/info/sync", s.Authorise(syncInfoHandler)).Methods(http.MethodPost)
		router.Handle("/host", s.Authorise(hostQueryHandler)).Methods(http.MethodGet)
		router.Handle("/host/{host-uuid}", s.Authorise(hostDecommissionHandler)).Methods(http.MethodDelete)
		router.Handle("/cmd", s.Authorise(updateCmdHandler)).Methods("PUT")
		router.Handle("/cmd", s.Authorise(getAllCmdHandler)).Methods(http.MethodGet)
		router.Handle("/cmd/{name}", s.Authorise(getCmdHandler)).Methods(http.MethodGet)
		router.Handle("/cmd/{name}", s.Authorise(deleteCmdHandler)).Methods(http.MethodDelete)
		router.Handle("/org-group", s.Authorise(getOrgGroupsHandler)).Methods(http.MethodGet)
		router.Handle("/org-group/{org-group}/area", s.Authorise(getAreasHandler)).Methods(http.MethodGet)
		router.Handle("/org-group/{org-group}/org", s.Authorise(getOrgHandler)).Methods(http.MethodGet)
		router.Handle("/area/{area}/location", s.Authorise(getLocationsHandler)).Methods(http.MethodGet)
		router.Handle("/admission", s.Authorise(setAdmissionHandler)).Methods(http.MethodPut)
		router.Handle("/package", s.Authorise(getPackagesHandler)).Methods(http.MethodGet, http.MethodOptions)
		router.Handle("/package/{name}/api", s.Authorise(getPackagesApiHandler)).Methods(http.MethodGet)
		router.Handle("/job", s.Authorise(newJobHandler)).Methods(http.MethodPost)
		router.Handle("/job", s.Authorise(getJobsHandler)).Methods(http.MethodGet)
		router.Handle("/job/batch", s.Authorise(getJobBatchHandler)).Methods(http.MethodGet)
		router.Handle("/user", s.Authorise(getUserHandler)).Methods(http.MethodGet)
		router.Handle("/dictionary/{key}", s.Authorise(getDictionaryHandler)).Methods(http.MethodGet)
		router.Handle("/dictionary", s.Authorise(setDictionaryHandler)).Methods(http.MethodPut)
		router.Handle("/dictionary/{key}", s.Authorise(deleteDictionaryHandler)).Methods(http.MethodDelete)
		router.Handle("/dictionary", s.Authorise(getDictionaryListHandler)).Methods(http.MethodGet)
		router.Handle("/cve/baseline", s.Authorise(getCVEBaselineHandler)).Methods(http.MethodGet)

		router.HandleFunc("/pub", getKeyHandler).Methods(http.MethodGet)

		router.HandleFunc("/activation/{macAddress}/{uuid}", activationHandler).Methods(http.MethodPost)
		router.HandleFunc("/registration", registrationHandler).Methods("POST")
		router.HandleFunc("/registration/{mac-address}", undoRegistrationHandler).Methods(http.MethodDelete)
	}
	// set up specific authentication for host pilot agents
	s.Auth = map[string]func(http.Request) (*h.UserPrincipal, error){
		"^/register":         pilotAuth,
		"^/ping":             pilotAuth,
		"^/cve/upload":       pilotAuth,
		"^/metrics/*":        pilotAuth,
		"^/logs/*":           pilotAuth,
		"^/activation/.*/.*": activationSvc,
		"^/pub":              nil,
		"^/$":                nil,
	}
	s.DefaultAuth = defaultAuth
	// s.Jobs = func() error {
	// 	enableTelemetry := os.Getenv("PILOTCTL_ENABLE_TELEMETRY")
	// 	if len(enableTelemetry) > 0 {
	// 		// launches the OT gateway
	// 		gateway := telem.NewOtServer(0)
	// 		go gateway.Start()
	// 	}
	// 	return nil
	// }
	s.Serve()
}

// the overridden authentication mechanism used by the authentication middleware for specific routes
// specified in server.Auth map
var pilotAuth = func(r http.Request) (*h.UserPrincipal, error) {
	token := r.Header.Get("Authorization")
	return core.Api().AuthenticatePilot(token)
}

// the default authentication mechanism user by the authentication middleware
var defaultAuth = func(r http.Request) (*h.UserPrincipal, error) {
	return core.Api().AuthenticateUser(r)
}

// authenticates requests from the activation service
var activationSvc = func(r http.Request) (*h.UserPrincipal, error) {
	return core.Api().AuthenticateActivationSvc(r)
}
