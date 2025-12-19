package config

import (
	"flag"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port             int
	DBPath           string
	AuthUsername     string
	AuthPassword     string
	LogRetentionDays int
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

	// Log retention (default 30 days)
	logRetention := 30
	if retentionEnv := os.Getenv("LOG_RETENTION_DAYS"); retentionEnv != "" {
		if parsed, err := strconv.Atoi(retentionEnv); err == nil && parsed > 0 {
			logRetention = parsed
		}
	}

	cfg := &Config{
		Port:             8080, // default HTTP port
		DBPath:           dbPath,
		AuthUsername:     authUser,
		AuthPassword:     authPass,
		LogRetentionDays: logRetention,
	}

	// Command-line flags
	port := fs.Int("port", cfg.Port, "HTTP server port")
	dbPathFlag := fs.String("db", cfg.DBPath, "Database file path")
	authUserFlag := fs.String("auth-user", cfg.AuthUsername, "Username for HTTP Basic Auth (optional)")
	authPassFlag := fs.String("auth-pass", cfg.AuthPassword, "Password for HTTP Basic Auth (optional)")
	logRetentionFlag := fs.Int("log-retention-days", cfg.LogRetentionDays, "Number of days to retain logs (0 = keep forever)")
	_ = fs.Parse(args)

	cfg.Port = *port
	cfg.DBPath = *dbPathFlag
	cfg.AuthUsername = *authUserFlag
	cfg.AuthPassword = *authPassFlag
	if *logRetentionFlag >= 0 {
		cfg.LogRetentionDays = *logRetentionFlag
	}

	return cfg
}
