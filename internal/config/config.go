package config

import (
	"flag"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port int
}

// Load loads configuration from flags and environment variables
// Priority: command-line flag > environment variable > default (8080)
func Load() *Config {
	return LoadWithFlagSet(flag.CommandLine, os.Args[1:])
}

// LoadWithFlagSet loads configuration with a custom flag set (for testing)
func LoadWithFlagSet(fs *flag.FlagSet, args []string) *Config {
	cfg := &Config{
		Port: 8080, // default port
	}

	// Check environment variable
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if port, err := strconv.Atoi(portEnv); err == nil {
			cfg.Port = port
		}
	}

	// Command-line flag has highest precedence
	port := fs.Int("port", cfg.Port, "HTTP server port")
	fs.Parse(args)
	cfg.Port = *port

	return cfg
}
