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

	publisher  platform.Publisher
	subscriber platform.Subscriber

	SSL_CERT_FILE = "/tmp/server.cert"
	SSL_KEY_FILE  = "/tmp/server.key"
)
