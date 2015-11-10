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

// EnforceHeadersHandler - Will enforce headers that are required by microplatform-io
func EnforceHeadersHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Add("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Add("Access-Control-Allow-Origin", "null")
		}

		w.Header().Add("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Connection", "keep-alive")

		next.ServeHTTP(w, r)
	})
}

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
			errResp, _ := json.Marshal(map[string]string{
				"message": fmt.Sprintf("Failed to read body: %s", err),
			})
			w.Write(errResp)
			return
		}

		platformRequestBytes, err := hex.DecodeString(fmt.Sprintf("%s", contents))

		if err != nil {
			errResp, _ := json.Marshal(map[string]string{
				"message": fmt.Sprintf("Failed to decode body: %s", err),
			})
			w.Write(errResp)
			return
		}

		platformRequest := &platform.Request{}

		if err := platform.Unmarshal(platformRequestBytes, platformRequest); err != nil {
			errResp, _ := json.Marshal(map[string]string{
				"message": fmt.Sprintf("Failed to unmarshal platform request: %s", err),
			})
			w.Write(errResp)
			return
		}

		logger.Print(platformRequest.GetUuid())

		if platformRequest.Routing == nil {
			platformRequest.Routing = &platform.Routing{}
		}

		if !platform.RouteToSchemeMatches(platformRequest, "microservice") {
			errResp, _ := json.Marshal(map[string]string{
				"message": fmt.Sprintf("Unsupported scheme provided: %s", platformRequest.Routing.RouteTo),
			})
			w.Write(errResp)
			return
		}

		responses, timeout := server.Router.Route(platformRequest)

		for {
			select {
			case response := <-responses:
				logger.Printf("{socket_id:'%s'} - got a response for request: %s", 12, platformRequest.GetUuid())
				logger.Printf("%q", response)

				/**
				response.Uuid = platform.String(strings.Replace(response.GetUuid(), requestUuidPrefix, "", -1))

				// Strip off the tail for routing
				response.Routing.RouteTo = response.Routing.RouteTo[:len(response.Routing.RouteTo)-1]

				responseBytes, err := platform.Marshal(response)
				if err != nil {
					log.Printf("[subscriber] failed to marshal platform request: %s", err)
					return
				}

				if err := so.Emit("request", hex.EncodeToString(responseBytes)); err != nil {
					log.Printf("[subscriber] failed to send platform request: %s", err)
					return
				}

				if response.GetCompleted() {
					log.Printf("{socket_id:'%s'} - got the final response for request: %s", socketId, platformRequest.GetUuid())
					return
				}

				**/

			case <-timeout:
				errResp, _ := json.Marshal(map[string]string{
					"message": fmt.Sprintf("Got a timeout for request: %s", platformRequest.GetUuid()),
				})
				w.Write(errResp)
				return
			}
		}

		return
	}
}
