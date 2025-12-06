package config

import (
	"flag"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port       int
	TLSEnabled bool
	TLSCert    string
	TLSKey     string
}

// Load loads configuration from flags and environment variables
// Priority: command-line flag > environment variable > default (8080)
func Load() *Config {
	return LoadWithFlagSet(flag.CommandLine, os.Args[1:])
}

// LoadWithFlagSet loads configuration with a custom flag set (for testing)
func LoadWithFlagSet(fs *flag.FlagSet, args []string) *Config {
	cfg := &Config{
		Port:       8080, // default port
		TLSEnabled: false,
		TLSCert:    "server.crt",
		TLSKey:     "server.key",
	}

	// Check environment variables
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if port, err := strconv.Atoi(portEnv); err == nil {
			cfg.Port = port
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

	// Command-line flags have highest precedence
	port := fs.Int("port", cfg.Port, "HTTP server port")
	tlsEnabled := fs.Bool("tls", cfg.TLSEnabled, "Enable TLS/HTTPS")
	tlsCert := fs.String("tls-cert", cfg.TLSCert, "Path to TLS certificate file")
	tlsKey := fs.String("tls-key", cfg.TLSKey, "Path to TLS private key file")
	fs.Parse(args)

	cfg.Port = *port
	cfg.TLSEnabled = *tlsEnabled
	cfg.TLSCert = *tlsCert
	cfg.TLSKey = *tlsKey

	return cfg
}
