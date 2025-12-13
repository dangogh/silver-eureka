package config

import (
	"flag"
	"os"
)

// Config holds the application configuration
type Config struct {
	Port         int
	DBPath       string
	AuthUsername string
	AuthPassword string
}

// Load loads configuration from flags
// Priority: command-line flag > environment variable > default
func Load() *Config {
	return LoadWithFlagSet(flag.CommandLine, os.Args[1:])
}

// LoadWithFlagSet loads configuration with a custom flag set (for testing)
func LoadWithFlagSet(fs *flag.FlagSet, args []string) *Config {
	// Check environment variable for database path
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/requests.db"
	}

	cfg := &Config{
		Port:   8080, // default HTTP port
		DBPath: dbPath,
	}

	// Command-line flags
	port := fs.Int("port", cfg.Port, "HTTP server port")
	dbPathFlag := fs.String("db", cfg.DBPath, "Database file path")
	fs.Parse(args)

	cfg.Port = *port
	cfg.DBPath = *dbPathFlag

	return cfg
}
