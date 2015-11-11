// Package main ...
package main

import (
	"net/http"

	"github.com/microplatform-io/platform"
)

func main() {
	serverIPAddr, err := platform.GetMyIp()

	if err != nil {
		logger.Fatalf("Could not resolve server IP address: %s", err)
	}

	server, _ := NewServer(&Options{
		IPAddr:      platform.Getenv("SERVER_IP", serverIPAddr),
		Port:        serverPort,
		TLSCertFile: certFile,
		TLSKeyFile:  keyFile,
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
