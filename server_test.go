// Package main ...
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.teltech.co/teltech/teltech-go.git"
	"github.com/microplatform-io/platform"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testIPAddr    = "127.0.0.1"
	testPort      = "43312"
	testPortMicro = "43313"
	testCookie    = "user=gxYzM4ZDg5MjdlODJkZDJlZWEifQ%3D%3D;"
	testToken     = "FFF90B88-EE71-3138-9CC4-A1998A3ABU91"
)

func getServer(port string) (*Server, error) {
	return NewServer(&Options{
		IPAddr:      testIPAddr,
		Port:        port,
		TLSCertFile: "test_data/server.pem",
		TLSKeyFile:  "test_data/server.key",
	})
}

func TestServerTLSInstance(t *testing.T) {
	server, err := getServer(testPort)

	Convey("By creating new server we get proper server instance and no errors", t, func() {
		So(*server, ShouldHaveSameTypeAs, Server{})
		So(err, ShouldBeNil)
	})
}

func TestServerDiscoveryHandler(t *testing.T) {
	server, _ := getServer(testPort)

	Convey("By accessing /server we should get back proper service descovery json", t, func() {

		// Server really does not need to be mocked at all
		mux := http.NewServeMux()
		mux.Handle("/server", EnforceHeadersMiddleware(http.HandlerFunc(ServerDiscoveryHandler(server))))

		go func() {
			err := server.ListenAndServeTLS(mux)
			Convey("No errors occurred while listening for server discovery platform endpoint @ /server", t, func() {
				So(err, ShouldBeNil)
			})
		}()

		// Give it time to start the listener
		time.Sleep(300 * time.Millisecond)

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		response, err := client.Get(fmt.Sprintf("https://%s/server", server.Addr))
		So(err, ShouldBeNil)

		contents, err := ioutil.ReadAll(response.Body)
		So(err, ShouldBeNil)

		shouldMatch := fmt.Sprintf(
			`{"host":"%s","port":"%s","protocol":"%s"}`,
			server.GetFormattedHostAddr(), server.Options.Port, serverProtocol,
		)
		So(string(contents), ShouldEqual, shouldMatch)
	})
}

func TestMicroplatformHandler(t *testing.T) {
	server, _ := getServer(testPortMicro)

	connectionManager := platform.NewAmqpConnectionManager(rabbitUser, rabbitPass, rabbitAddr+":"+rabbitPort, "")
	publisher = getDefaultPublisher(connectionManager)
	subscriber = getDefaultSubscriber(connectionManager, server.GetRouterURI())

	router := platform.NewStandardRouter(publisher, subscriber)
	server.SetRouter(router)

	Convey("By accessing / with hex encoded string we should get back proper endpoint response", t, func() {

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "text/plain")

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

			if platformRequest.Routing.RouteFrom != nil {
				platformRequest.Routing.RouteFrom = []*platform.Route{}
			}

			if !platform.RouteToSchemeMatches(platformRequest, "microservice") {
				w.Write(ErrorResponse(fmt.Sprintf("Unsupported scheme provided: %s", platformRequest.Routing.RouteTo)))
				return
			}

			w.Header().Set("Content-Type", "text/plain")

			mainReq := platform.Request{
				Routing: &platform.Routing{
					RouteTo: []*platform.Route{&platform.Route{Uri: teltech.String("resource:///testing/reply/http-router")}},
				},
			}

			w.Write([]byte(hex.EncodeToString(GetProtoBytes(platform.GenerateResponse(&mainReq, &mainReq)))))
			return
		}))

		defer server.Close()

		// Give it time to start the listener
		time.Sleep(300 * time.Millisecond)

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		Convey("By passing valid / we should get back valid response", func() {
			req := &platform.Request{
				Routing: &platform.Routing{
					RouteTo: []*platform.Route{&platform.Route{Uri: teltech.String("microservice:///testing/get/http-router")}},
				},
				Completed: teltech.Bool(true),
			}

			responseBytes, err := platform.Marshal(req)
			So(err, ShouldBeNil)

			reqHex := []byte(hex.EncodeToString(responseBytes))
			request, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(reqHex))
			So(err, ShouldBeNil)

			request.Header.Set("Content-Type", "text/plain")

			resp, err := client.Do(request)
			So(err, ShouldBeNil)

			defer resp.Body.Close()

			contents, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)

			platformResponseBytes, err := hex.DecodeString(string(contents))
			So(err, ShouldBeNil)

			response := &platform.Request{}
			err = platform.Unmarshal(platformResponseBytes, response)
			So(err, ShouldBeNil)

			So(response.GetRouting().GetRouteTo()[0].GetUri(), ShouldNotEqual, "resource:///meta/reply/error")
			So(response.GetRouting().GetRouteTo()[0].GetUri(), ShouldEqual, "resource:///testing/reply/http-router")
		})

		Convey("By passing invalid hex message towards / we should get back hex decode errors", func() {
			reqHex := []byte("I am not valid request")
			request, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(reqHex))
			So(err, ShouldBeNil)

			request.Header.Set("Content-Type", "text/plain")

			resp, err := client.Do(request)
			So(err, ShouldBeNil)

			defer resp.Body.Close()

			contents, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)

			platformResponseBytes, err := hex.DecodeString(string(contents))
			So(err, ShouldBeNil)

			response := &platform.Request{}
			err = platform.Unmarshal(platformResponseBytes, response)
			So(err, ShouldBeNil)

			So(response.GetRouting().GetRouteTo()[0].GetUri(), ShouldNotEqual, "resource:///meta/reply/error")

			responseErr := &platform.Error{}
			err = platform.Unmarshal(response.GetPayload(), responseErr)
			So(err, ShouldBeNil)
			So(responseErr.GetMessage(), ShouldContainSubstring, "Failed to decode body")
		})

		Convey("By passing invalid hex request towards / we should get back unmarshal errors", func() {
			reqHex := []byte(hex.EncodeToString([]byte("Somewhere over the rainbow...")))
			request, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(reqHex))
			So(err, ShouldBeNil)

			request.Header.Set("Content-Type", "text/plain")

			resp, err := client.Do(request)
			So(err, ShouldBeNil)

			defer resp.Body.Close()

			contents, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)

			platformResponseBytes, err := hex.DecodeString(string(contents))
			So(err, ShouldBeNil)

			response := &platform.Request{}
			err = platform.Unmarshal(platformResponseBytes, response)
			So(err, ShouldBeNil)

			So(response.GetRouting().GetRouteTo()[0].GetUri(), ShouldNotEqual, "resource:///meta/reply/error")

			responseErr := &platform.Error{}
			err = platform.Unmarshal(response.GetPayload(), responseErr)
			So(err, ShouldBeNil)
			So(responseErr.GetMessage(), ShouldContainSubstring, "Failed to unmarshal platform request")
		})

		Convey("By passing valid hex and request but invalid router scheme we should see invald scheme error", func() {
			req := &platform.Request{
				Routing: &platform.Routing{
					RouteTo: []*platform.Route{&platform.Route{Uri: teltech.String("microservice-invalid:///testing/get/http-router")}},
				},
				Completed: teltech.Bool(true),
			}

			responseBytes, err := platform.Marshal(req)
			So(err, ShouldBeNil)

			reqHex := []byte(hex.EncodeToString(responseBytes))
			request, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(reqHex))
			So(err, ShouldBeNil)

			request.Header.Set("Content-Type", "text/plain")

			resp, err := client.Do(request)
			So(err, ShouldBeNil)

			defer resp.Body.Close()

			contents, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)

			platformResponseBytes, err := hex.DecodeString(string(contents))
			So(err, ShouldBeNil)

			response := &platform.Request{}
			err = platform.Unmarshal(platformResponseBytes, response)
			So(err, ShouldBeNil)

			So(response.GetRouting().GetRouteTo()[0].GetUri(), ShouldNotEqual, "resource:///meta/reply/error")

			responseErr := &platform.Error{}
			err = platform.Unmarshal(response.GetPayload(), responseErr)
			So(err, ShouldBeNil)
			So(responseErr.GetMessage(), ShouldContainSubstring, "Unsupported scheme provided")
		})
	})
}
