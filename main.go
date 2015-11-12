// Package main ...
package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/microplatform-io/platform"
)

func main() {
	serverIPAddr, err := platform.GetMyIp()

	if err != nil {
		logger.Fatalf("Could not resolve server IP address: %s", err)
	}

	if err := ioutil.WriteFile(SSL_CERT_FILE, []byte(strings.Replace(os.Getenv("SSL_CERT"), "\\n", "\n", -1)), 0755); err != nil {
		log.Fatalf("> failed to write SSL cert file: %s", err)
	}

	if err := ioutil.WriteFile(SSL_KEY_FILE, []byte(strings.Replace(os.Getenv("SSL_KEY"), "\\n", "\n", -1)), 0755); err != nil {
		log.Fatalf("> failed to write SSL cert file: %s", err)
	}

	server, _ := NewServer(&Options{
		IPAddr:      platform.Getenv("SERVER_IP", serverIPAddr),
		MicroIpAddr: serverIPAddr,
		Port:        serverPort,
		TLSCertFile: SSL_CERT_FILE,
		TLSKeyFile:  SSL_KEY_FILE,
	})

	connectionManager := platform.NewAmqpConnectionManager(rabbitUser, rabbitPass, rabbitAddr+":"+rabbitPort, "")
	publisher = getDefaultPublisher(connectionManager)
	subscriber = getDefaultSubscriber(connectionManager, server.GetRouterURI())

	router := platform.NewStandardRouter(publisher, subscriber)
	server.SetRouter(router)

	manageRouterState(&platform.RouterConfigList{
		RouterConfigs: []*platform.RouterConfig{
			&platform.RouterConfig{
				RouterType:   platform.RouterConfig_ROUTER_TYPE_HTTP.Enum(),
				ProtocolType: platform.RouterConfig_PROTOCOL_TYPE_HTTPS.Enum(),
				Host:         platform.String(serverIPAddr),
				Port:         platform.String(serverPort),
			},
		},
	})

	mux := http.NewServeMux()

	mux.Handle("/", EnforceHeadersMiddleware(http.HandlerFunc(MicroplatformEndpointHandler(server))))
	mux.Handle("/server", EnforceHeadersMiddleware(http.HandlerFunc(ServerDiscoveryHandler(server))))

	if err := server.ListenAndServeTLS(mux); err != nil {
		logger.Fatalf("Failed to listen and serve: %s", err)
	}

}
