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
	if cfg.LogRetentionDays != 30 {
		t.Errorf("Expected default log retention 30 days, got %d", cfg.LogRetentionDays)
	}
}

func TestLoad_CommandLineFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{"-port=3000", "-db=/tmp/test.db", "-log-retention-days=60"})

	if cfg.Port != 3000 {
		t.Errorf("Expected port 3000 from flag, got %d", cfg.Port)
	}
	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("Expected db path '/tmp/test.db' from flag, got %s", cfg.DBPath)
	}
	if cfg.LogRetentionDays != 60 {
		t.Errorf("Expected log retention 60 days from flag, got %d", cfg.LogRetentionDays)
	}
}

func TestLoad_LogRetentionZero(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{"-log-retention-days=0"})

	if cfg.LogRetentionDays != 0 {
		t.Errorf("Expected log retention 0 (disabled), got %d", cfg.LogRetentionDays)
	}
}
