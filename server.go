// Package main ...
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/microplatform-io/platform"
)

// Server - Wrapper on top of router with some additional DRY features to keep things
// clear and consistent.
type Server struct {
	*Options

	Addr string
	platform.Router
}

// SetRouter - Shortcut for setting up platform router.
func (s *Server) SetRouter(router platform.Router) {
	s.Router = router
}

// GetHostname - Will return hostname of current system.
func (s *Server) GetHostname() (hostname string) {
	hostname, _ = os.Hostname()
	return
}

// GetRouterURI - Will return router uri required by micro platform.
func (s *Server) GetRouterURI() string {
	return fmt.Sprintf("router-%s", s.GetHostname())
}

// GetFormattedHostAddr - Will return formatted host address that is required by microplatform.
func (s *Server) GetFormattedHostAddr() string {
	hostAddress := strings.Replace(s.Options.MicroIpAddr, ".", "-", -1)
	return fmt.Sprintf("%s.%s", hostAddress, "microplatform.io")
}

// ListenAndServeTLS - Will start and listen HTTP/TLS server
func (s *Server) ListenAndServeTLS(mux *http.ServeMux) error {
	logger.Printf("Starting service TLS server on: %s", s.Addr)
	return http.ListenAndServeTLS(s.Addr, s.Options.TLSCertFile, s.Options.TLSKeyFile, mux)
}

// NewServer - Will return back new instance of server. Server is used just as a wrapper
// to make things more organized.
func NewServer(o *Options) (*Server, error) {
	return &Server{Addr: fmt.Sprintf("%s:%s", o.IPAddr, o.Port), Options: o}, nil
}
