// Package main ...
package main

import (
	"os"

	"github.com/microplatform-io/platform"
)

var (
	rabbitUser     = os.Getenv("RABBITMQ_USER")
	rabbitPass     = os.Getenv("RABBITMQ_PASS")
	rabbitAddr     = os.Getenv("RABBITMQ_PORT_5672_TCP_ADDR")
	rabbitPort     = os.Getenv("RABBITMQ_PORT_5672_TCP_PORT")
	serverProtocol = platform.Getenv("SERVER_PROTOCOL", "https")
	serverPort     = platform.Getenv("PORT", "443")
	certFile       = os.Getenv("SSL_CERT")
	keyFile        = os.Getenv("SSL_KEY")

	publisher  platform.Publisher
	subscriber platform.Subscriber
)
