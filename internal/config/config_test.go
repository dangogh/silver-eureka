package config

import (
	"flag"
	"os"
	"testing"
)

func TestLoad_DefaultPort(t *testing.T) {
	os.Unsetenv("PORT")

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{})

	if cfg.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Port)
	}
}

func TestLoad_EnvironmentVariable(t *testing.T) {
	os.Setenv("PORT", "9090")
	defer os.Unsetenv("PORT")

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{})

	if cfg.Port != 9090 {
		t.Errorf("Expected port 9090 from environment, got %d", cfg.Port)
	}
}

func TestLoad_CommandLineFlag(t *testing.T) {
	// Set environment variable to verify flag takes precedence
	os.Setenv("PORT", "9090")
	defer os.Unsetenv("PORT")

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{"-port=7070"})

	if cfg.Port != 7070 {
		t.Errorf("Expected port 7070 from flag (highest precedence), got %d", cfg.Port)
	}
}

func TestLoad_InvalidEnvironmentVariable(t *testing.T) {
	os.Setenv("PORT", "invalid")
	defer os.Unsetenv("PORT")

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{})

	// Should fall back to default
	if cfg.Port != 8080 {
		t.Errorf("Expected default port 8080 when env is invalid, got %d", cfg.Port)
	}
}
