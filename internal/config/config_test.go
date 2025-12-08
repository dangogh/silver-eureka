package config

import (
	"flag"
	"testing"
)

func TestLoad_DefaultPort(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{})

	if cfg.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Port)
	}
}

func TestLoad_CommandLineFlag(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{"-port=3000"})

	if cfg.Port != 3000 {
		t.Errorf("Expected port 3000 from flag, got %d", cfg.Port)
	}
}
