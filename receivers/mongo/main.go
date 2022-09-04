/*
  pilot control service - mongo event receiver
  © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package main

import (
	"github.com/gorilla/mux"
	h "southwinds.dev/http"
	"southwinds.dev/pilotctl/receivers/mongo/core"
)

func main() {
	// creates a generic http server
	s := h.New("pilotctl-mongo-receiver", core.Version)
	s.Http = func(router *mux.Router) {
		// enable encoded path  vars
		router.UseEncodedPath()
		// middleware
		// router.Use(s.LoggingMiddleware)
		router.Use(s.AuthenticationMiddleware)
		// add http handlers
		router.HandleFunc("/events", eventReceiverHandler).Methods("POST")
		router.HandleFunc("/events", eventQueryHandler).Methods("GET")
	}
	s.Serve()
}
