package config

import (
	"flag"
	"os"
)

// Config holds the application configuration
type Config struct {
	Port int
}

// Load loads configuration from flags
// Priority: command-line flag > environment variable > default
func Load() *Config {
	return LoadWithFlagSet(flag.CommandLine, os.Args[1:])
}

// LoadWithFlagSet loads configuration with a custom flag set (for testing)
func LoadWithFlagSet(fs *flag.FlagSet, args []string) *Config {
	cfg := &Config{
		Port: 8080, // default HTTP port
	}

	// Command-line flag for port override
	port := fs.Int("port", cfg.Port, "HTTP server port")
	fs.Parse(args)

	cfg.Port = *port

	return cfg
}
