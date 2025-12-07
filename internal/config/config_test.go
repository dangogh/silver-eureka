package config

import (
	"flag"
	"testing"
)

func TestLoad_DefaultPort(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{})

	if cfg.Port != 443 {
		t.Errorf("Expected default port 443, got %d", cfg.Port)
	}
	if cfg.TLSCert != "server.crt" {
		t.Errorf("Expected default TLS cert 'server.crt', got %s", cfg.TLSCert)
	}
	if cfg.TLSKey != "server.key" {
		t.Errorf("Expected default TLS key 'server.key', got %s", cfg.TLSKey)
	}
}

func TestLoad_CommandLineFlag(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{"-port=8443"})

	if cfg.Port != 8443 {
		t.Errorf("Expected port 8443 from flag, got %d", cfg.Port)
	}
}
