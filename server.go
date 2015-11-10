// Package main ...
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/microplatform-io/platform"
)

// Server -
type Server struct {
	*Options

	Addr string
	platform.Router
}

// SetRouter -
func (s *Server) SetRouter(router platform.Router) {
	s.Router = router
}

// GetHostname -
func (s *Server) GetHostname() (hostname string) {
	hostname, _ = os.Hostname()
	return
}

// GetRouterURI -
func (s *Server) GetRouterURI() string {
	return fmt.Sprintf("router-%s", s.GetHostname())
}

// GetFormattedHostAddr -
func (s *Server) GetFormattedHostAddr() string {
	hostAddress := strings.Replace(s.Options.IPAddr, ".", "-", -1)
	return fmt.Sprintf("%s.%s", hostAddress, "microplatform.io")
}

// ListenAndServeTLS -
func (s *Server) ListenAndServeTLS(mux *http.ServeMux) error {
	logger.Printf("Starting service TLS server on: %s", s.GetFormattedHostAddr())
	return http.ListenAndServeTLS(s.Addr, s.Options.TLSCertFile, s.Options.TLSKeyFile, mux)
}

// NewServer -
func NewServer(o *Options) (*Server, error) {
	return &Server{
		Addr:    fmt.Sprintf("%s:%s", o.IPAddr, o.Port),
		Options: o,
	}, nil
}
