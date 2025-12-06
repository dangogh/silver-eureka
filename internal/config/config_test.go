package config

import (
	"flag"
	"os"
	"testing"
)

func TestLoad_DefaultPort(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("HTTP_PORT")
	os.Unsetenv("TLS_ENABLED")
	os.Unsetenv("TLS_CERT")
	os.Unsetenv("TLS_KEY")
	os.Unsetenv("HTTP_REDIRECT")

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{})

	if cfg.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Port)
	}
	if cfg.HTTPPort != 8000 {
		t.Errorf("Expected default HTTP port 8000, got %d", cfg.HTTPPort)
	}
	if cfg.TLSEnabled {
		t.Errorf("Expected TLS disabled by default")
	}
	if cfg.TLSCert != "server.crt" {
		t.Errorf("Expected default TLS cert 'server.crt', got %s", cfg.TLSCert)
	}
	if cfg.TLSKey != "server.key" {
		t.Errorf("Expected default TLS key 'server.key', got %s", cfg.TLSKey)
	}
	if !cfg.HTTPRedirect {
		t.Errorf("Expected HTTP redirect enabled by default")
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

func TestLoad_TLS_EnvironmentVariables(t *testing.T) {
	os.Setenv("TLS_ENABLED", "true")
	os.Setenv("TLS_CERT", "/path/to/cert.pem")
	os.Setenv("TLS_KEY", "/path/to/key.pem")
	defer func() {
		os.Unsetenv("TLS_ENABLED")
		os.Unsetenv("TLS_CERT")
		os.Unsetenv("TLS_KEY")
	}()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{})

	if !cfg.TLSEnabled {
		t.Errorf("Expected TLS enabled from environment")
	}
	if cfg.TLSCert != "/path/to/cert.pem" {
		t.Errorf("Expected TLS cert '/path/to/cert.pem', got %s", cfg.TLSCert)
	}
	if cfg.TLSKey != "/path/to/key.pem" {
		t.Errorf("Expected TLS key '/path/to/key.pem', got %s", cfg.TLSKey)
	}
}

func TestLoad_TLS_CommandLineFlags(t *testing.T) {
	// Set environment variables to verify flags take precedence
	os.Setenv("TLS_ENABLED", "false")
	os.Setenv("TLS_CERT", "/env/cert.pem")
	os.Setenv("TLS_KEY", "/env/key.pem")
	defer func() {
		os.Unsetenv("TLS_ENABLED")
		os.Unsetenv("TLS_CERT")
		os.Unsetenv("TLS_KEY")
	}()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{
		"-tls=true",
		"-tls-cert=/flag/cert.pem",
		"-tls-key=/flag/key.pem",
	})

	if !cfg.TLSEnabled {
		t.Errorf("Expected TLS enabled from flag (highest precedence)")
	}
	if cfg.TLSCert != "/flag/cert.pem" {
		t.Errorf("Expected TLS cert '/flag/cert.pem' from flag, got %s", cfg.TLSCert)
	}
	if cfg.TLSKey != "/flag/key.pem" {
		t.Errorf("Expected TLS key '/flag/key.pem' from flag, got %s", cfg.TLSKey)
	}
}

func TestLoad_HTTPRedirect_EnvironmentVariables(t *testing.T) {
	os.Setenv("HTTP_PORT", "9000")
	os.Setenv("HTTP_REDIRECT", "false")
	defer func() {
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("HTTP_REDIRECT")
	}()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{})

	if cfg.HTTPPort != 9000 {
		t.Errorf("Expected HTTP port 9000 from environment, got %d", cfg.HTTPPort)
	}
	if cfg.HTTPRedirect {
		t.Errorf("Expected HTTP redirect disabled from environment")
	}
}

func TestLoad_HTTPRedirect_CommandLineFlags(t *testing.T) {
	os.Setenv("HTTP_PORT", "9000")
	os.Setenv("HTTP_REDIRECT", "true")
	defer func() {
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("HTTP_REDIRECT")
	}()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := LoadWithFlagSet(fs, []string{
		"-http-port=7000",
		"-http-redirect=false",
	})

	if cfg.HTTPPort != 7000 {
		t.Errorf("Expected HTTP port 7000 from flag (highest precedence), got %d", cfg.HTTPPort)
	}
	if cfg.HTTPRedirect {
		t.Errorf("Expected HTTP redirect disabled from flag (highest precedence)")
	}
}
