// Package main ...
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
		mux.Handle("/server", EnforceHeadersHandler(http.HandlerFunc(ServerDiscoveryHandler(server))))

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

	Convey("By accessing /server we should get back proper service descovery json 2", t, func() {

		// Server really does not need to be mocked at all
		mux := http.NewServeMux()
		mux.Handle("/", EnforceHeadersHandler(http.HandlerFunc(MicroplatformEndpointHandler(server))))

		go func() {
			err := server.ListenAndServeTLS(mux)
			Convey("No errors occurred while listening for microplatform endpoint discovery @ /", t, func() {
				So(err, ShouldBeNil)
			})
		}()

		// Give it time to start the listener
		time.Sleep(300 * time.Millisecond)

		req := &platform.Request{
			Routing: &platform.Routing{
				RouteTo: []*platform.Route{
					&platform.Route{
						Uri: teltech.String("microservice://meta/get/truecaller"),
					},
				},
				RouteFrom: []*platform.Route{
					&platform.Route{
						Uri: teltech.String("microservice://testing/get/truecaller"),
					},
				},
			},
			Payload:   []byte(`{"phone_number": "+385915256970", "cookie": "XYZ"}`),
			Completed: teltech.Bool(true),
		}

		responseBytes, err := platform.Marshal(req)

		if err != nil {
			log.Printf("[subscriber] failed to marshal platform request: %s", err)
			return
		}

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		reqHex := []byte(hex.EncodeToString(responseBytes))
		request, err := http.NewRequest("POST", fmt.Sprintf("https://%s/meta/get/truecaller", server.Addr), bytes.NewBuffer(reqHex))
		So(err, ShouldBeNil)

		request.Header.Set("Content-Type", "text/plain")

		resp, err := client.Do(request)
		So(err, ShouldBeNil)

		defer resp.Body.Close()

		contents, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldBeNil)

		fmt.Printf("Got shit back: %s", contents)

		time.Sleep(5 * time.Second)
		/**
		shouldMatch := fmt.Sprintf(
			`{"host":"%s","port":"%s","protocol":"%s"}`,
			server.GetFormattedHostAddr(), server.Options.Port, serverProtocol,
		)
		So(string(contents), ShouldEqual, shouldMatch)
		**/

	})
}
