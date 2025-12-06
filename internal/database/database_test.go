package database

import (
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Use temporary database file
	dbPath := "/tmp/test_requests.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("Expected non-nil database")
	}
}

func TestLogRequest(t *testing.T) {
	dbPath := "/tmp/test_requests_log.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Log a request
	err = db.LogRequest("192.168.1.1", "/test/path")
	if err != nil {
		t.Errorf("Failed to log request: %v", err)
	}

	// Verify the log was created
	logs, err := db.GetLogs(10)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(logs))
	}

	if logs[0].IPAddress != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", logs[0].IPAddress)
	}

	if logs[0].URL != "/test/path" {
		t.Errorf("Expected URL /test/path, got %s", logs[0].URL)
	}
}

func TestGetLogs(t *testing.T) {
	dbPath := "/tmp/test_requests_getlogs.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Log multiple requests
	testData := []struct {
		ip  string
		url string
	}{
		{"192.168.1.1", "/path1"},
		{"192.168.1.2", "/path2"},
		{"192.168.1.3", "/path3"},
	}

	for _, td := range testData {
		if err := db.LogRequest(td.ip, td.url); err != nil {
			t.Fatalf("Failed to log request: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get all logs
	logs, err := db.GetLogs(0)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(logs))
	}

	// Test with limit
	logs, err = db.GetLogs(2)
	if err != nil {
		t.Fatalf("Failed to get logs with limit: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 log entries with limit, got %d", len(logs))
	}

	// Verify logs are ordered by timestamp DESC (most recent first)
	if logs[0].URL != "/path3" {
		t.Errorf("Expected most recent log first (/path3), got %s", logs[0].URL)
	}
}

func TestClose(t *testing.T) {
	dbPath := "/tmp/test_requests_close.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Failed to close database: %v", err)
	}

	// Closing again should not panic
	err = db.Close()
	if err != nil {
		t.Errorf("Closing already closed database returned error: %v", err)
	}
}
