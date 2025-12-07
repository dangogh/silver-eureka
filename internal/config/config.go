package config

import (
	"flag"
	"os"
)

// Config holds the application configuration
type Config struct {
	Port    int
	TLSCert string
	TLSKey  string
}

// Load loads configuration from flags
// Priority: command-line flag > default (443)
func Load() *Config {
	return LoadWithFlagSet(flag.CommandLine, os.Args[1:])
}

// LoadWithFlagSet loads configuration with a custom flag set (for testing)
func LoadWithFlagSet(fs *flag.FlagSet, args []string) *Config {
	cfg := &Config{
		Port:    443, // default HTTPS port
		TLSCert: "server.crt",
		TLSKey:  "server.key",
	}

	// Command-line flag for port override
	port := fs.Int("port", cfg.Port, "HTTPS server port")
	fs.Parse(args)

	cfg.Port = *port

	return cfg
}
