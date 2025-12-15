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
	// Check environment variables
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/requests.db"
	}

	authUser := os.Getenv("AUTH_USERNAME")
	authPass := os.Getenv("AUTH_PASSWORD")

	cfg := &Config{
		Port:         8080, // default HTTP port
		DBPath:       dbPath,
		AuthUsername: authUser,
		AuthPassword: authPass,
	}

	// Command-line flags
	port := fs.Int("port", cfg.Port, "HTTP server port")
	dbPathFlag := fs.String("db", cfg.DBPath, "Database file path")
	authUserFlag := fs.String("auth-user", cfg.AuthUsername, "Username for HTTP Basic Auth (optional)")
	authPassFlag := fs.String("auth-pass", cfg.AuthPassword, "Password for HTTP Basic Auth (optional)")
	_ = fs.Parse(args)

	cfg.Port = *port
	cfg.DBPath = *dbPathFlag
	cfg.AuthUsername = *authUserFlag
	cfg.AuthPassword = *authPassFlag

	return cfg
}
