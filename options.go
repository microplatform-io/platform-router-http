// Package main ...
package main

// Options - Used as server options. Should be passed along with NewServer()
type Options struct {
	IPAddr      string
	Port        string
	TLSCertFile string
	TLSKeyFile  string
}
