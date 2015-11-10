// Package main ...
package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/microplatform-io/platform"
)

// ServerDiscoveryHandler - Will return (based on content type) service discovery details
func ServerDiscoveryHandler(server *Server) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		cb := req.FormValue("callback")

		jsonBytes, _ := json.Marshal(map[string]string{
			"protocol": serverProtocol,
			"host":     server.GetFormattedHostAddr(),
			"port":     server.Options.Port,
		})

		if cb == "" {
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonBytes)
			return
		}

		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(w, fmt.Sprintf("%s(%s)", cb, jsonBytes))
	}
}

// MicroplatformEndpointHandler -
func MicroplatformEndpointHandler(server *Server) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		contents, err := ioutil.ReadAll(req.Body)

		if err != nil {
			w.Write(ErrorResponse(fmt.Sprintf("Failed to read body: %s", err)))
			return
		}

		platformRequestBytes, err := hex.DecodeString(fmt.Sprintf("%s", contents))

		if err != nil {
			w.Write(ErrorResponse(fmt.Sprintf("Failed to decode body: %s", err)))
			return
		}

		platformRequest := &platform.Request{}

		if err := platform.Unmarshal(platformRequestBytes, platformRequest); err != nil {
			w.Write(ErrorResponse(fmt.Sprintf("Failed to unmarshal platform request: %s", err)))
			return
		}

		if platformRequest.Routing == nil {
			platformRequest.Routing = &platform.Routing{}
		}

		if !platform.RouteToSchemeMatches(platformRequest, "microservice") {
			w.Write(ErrorResponse(fmt.Sprintf("Unsupported scheme provided: %s", platformRequest.Routing.RouteTo)))
			return
		}

		responses, timeout := server.Router.Route(platformRequest)

		for {
			select {
			case response := <-responses:
				logger.Printf("Got a response for request:", platformRequest.GetUuid())

				responseBytes, err := platform.Marshal(response)

				if err != nil {
					w.Write(ErrorResponse(fmt.Sprintf(
						"failed to marshal platform request: %s - err: %s", platformRequest.GetUuid(), err,
					)))
					return
				}

				w.Write(responseBytes)

			case <-timeout:
				w.Write(ErrorResponse(fmt.Sprintf("Got a timeout for request: %s", platformRequest.GetUuid())))
				return
			}
		}

		return
	}
}
