package config

import (
	"flag"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port         int
	HTTPPort     int
	TLSEnabled   bool
	TLSCert      string
	TLSKey       string
	HTTPRedirect bool
}

// Load loads configuration from flags and environment variables
// Priority: command-line flag > environment variable > default (8080)
func Load() *Config {
	return LoadWithFlagSet(flag.CommandLine, os.Args[1:])
}

// LoadWithFlagSet loads configuration with a custom flag set (for testing)
func LoadWithFlagSet(fs *flag.FlagSet, args []string) *Config {
	cfg := &Config{
		Port:         8080, // default HTTPS port
		HTTPPort:     8000, // default HTTP port for redirect
		TLSEnabled:   false,
		TLSCert:      "server.crt",
		TLSKey:       "server.key",
		HTTPRedirect: true, // default to redirecting HTTP to HTTPS when TLS is enabled
	}

	// Check environment variables
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if port, err := strconv.Atoi(portEnv); err == nil {
			cfg.Port = port
		}
	}

	if httpPortEnv := os.Getenv("HTTP_PORT"); httpPortEnv != "" {
		if port, err := strconv.Atoi(httpPortEnv); err == nil {
			cfg.HTTPPort = port
		}
	}

	if tlsEnabled := os.Getenv("TLS_ENABLED"); tlsEnabled == "true" || tlsEnabled == "1" {
		cfg.TLSEnabled = true
	}

	if tlsCert := os.Getenv("TLS_CERT"); tlsCert != "" {
		cfg.TLSCert = tlsCert
	}

	if tlsKey := os.Getenv("TLS_KEY"); tlsKey != "" {
		cfg.TLSKey = tlsKey
	}

	if httpRedirect := os.Getenv("HTTP_REDIRECT"); httpRedirect == "false" || httpRedirect == "0" {
		cfg.HTTPRedirect = false
	}

	// Command-line flags have highest precedence
	port := fs.Int("port", cfg.Port, "HTTPS server port (or HTTP port if TLS disabled)")
	httpPort := fs.Int("http-port", cfg.HTTPPort, "HTTP redirect server port (when TLS enabled)")
	tlsEnabled := fs.Bool("tls", cfg.TLSEnabled, "Enable TLS/HTTPS")
	tlsCert := fs.String("tls-cert", cfg.TLSCert, "Path to TLS certificate file")
	tlsKey := fs.String("tls-key", cfg.TLSKey, "Path to TLS private key file")
	httpRedirect := fs.Bool("http-redirect", cfg.HTTPRedirect, "Enable HTTP to HTTPS redirect when TLS is enabled")
	fs.Parse(args)

	cfg.Port = *port
	cfg.HTTPPort = *httpPort
	cfg.TLSEnabled = *tlsEnabled
	cfg.TLSCert = *tlsCert
	cfg.TLSKey = *tlsKey
	cfg.HTTPRedirect = *httpRedirect

	return cfg
}
