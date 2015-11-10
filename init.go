// Package main ...
package main

import (
	"log"

	"github.com/microplatform-io/platform"
)

var (
	logger *log.Logger
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	logger = platform.GetLogger("platform-router-http")
}
