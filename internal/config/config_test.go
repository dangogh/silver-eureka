package config

import (
	"flag"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{})

	if cfg.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Port)
	}
	if cfg.DBPath != "data/requests.db" {
		t.Errorf("Expected default db path 'data/requests.db', got %s", cfg.DBPath)
	}
}

func TestLoad_CommandLineFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{"-port=3000", "-db=/tmp/test.db"})

	if cfg.Port != 3000 {
		t.Errorf("Expected port 3000 from flag, got %d", cfg.Port)
	}
	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("Expected db path '/tmp/test.db' from flag, got %s", cfg.DBPath)
	}
}
