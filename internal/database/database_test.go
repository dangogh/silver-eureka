package database

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Use temporary database file
	dbPath := "/tmp/test_requests.db"
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	if db == nil {
		t.Fatal("Expected non-nil database")
	}
}

func TestLogRequest(t *testing.T) {
	dbPath := "/tmp/test_requests_log.db"
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

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
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

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
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

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

func TestPing(t *testing.T) {
	dbPath := "/tmp/test_requests_ping.db"
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	// Test ping on open connection
	if err := db.Ping(); err != nil {
		t.Errorf("Ping failed on open connection: %v", err)
	}

	// Test ping after close
	if err := db.Close(); err != nil {
		t.Logf("Close error (expected): %v", err)
	}
	if err := db.Ping(); err == nil {
		t.Error("Expected error pinging closed database")
	}

	// Test ping with nil connection
	db.conn = nil
	if err := db.Ping(); err == nil {
		t.Error("Expected error pinging nil connection")
	}
}

func TestGetAllLogs(t *testing.T) {
	dbPath := "/tmp/test_requests_getall.db"
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	// Log multiple requests
	for i := 0; i < 5; i++ {
		if err := db.LogRequest("192.168.1.1", "/test"); err != nil {
			t.Fatalf("Failed to log request: %v", err)
		}
	}

	// Get all logs
	logs, err := db.GetAllLogs()
	if err != nil {
		t.Fatalf("Failed to get all logs: %v", err)
	}

	if len(logs) != 5 {
		t.Errorf("Expected 5 logs, got %d", len(logs))
	}
}

func TestGetEndpointStats(t *testing.T) {
	dbPath := "/tmp/test_requests_endpoint_stats.db"
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	// Log requests from multiple IPs to different endpoints
	testData := []struct {
		ip  string
		url string
	}{
		{"192.168.1.1", "/api/users"},
		{"192.168.1.2", "/api/users"},
		{"192.168.1.3", "/api/users"},
		{"192.168.1.1", "/api/posts"},
		{"192.168.1.2", "/api/posts"},
		{"192.168.1.1", "/health"},
	}

	for _, td := range testData {
		if err := db.LogRequest(td.ip, td.url); err != nil {
			t.Fatalf("Failed to log request: %v", err)
		}
	}

	// Get endpoint stats
	stats, err := db.GetEndpointStats()
	if err != nil {
		t.Fatalf("Failed to get endpoint stats: %v", err)
	}

	if len(stats) != 3 {
		t.Errorf("Expected 3 endpoints, got %d", len(stats))
	}

	// Verify stats are ordered by count DESC
	if stats[0].URL != "/api/users" {
		t.Errorf("Expected /api/users first, got %s", stats[0].URL)
	}
	if stats[0].Count != 3 {
		t.Errorf("Expected count 3 for /api/users, got %d", stats[0].Count)
	}
	if stats[0].UniqueIPs != 3 {
		t.Errorf("Expected 3 unique IPs for /api/users, got %d", stats[0].UniqueIPs)
	}

	if stats[1].URL != "/api/posts" {
		t.Errorf("Expected /api/posts second, got %s", stats[1].URL)
	}
	if stats[1].Count != 2 {
		t.Errorf("Expected count 2 for /api/posts, got %d", stats[1].Count)
	}
	if stats[1].UniqueIPs != 2 {
		t.Errorf("Expected 2 unique IPs for /api/posts, got %d", stats[1].UniqueIPs)
	}

	// Verify timestamps are populated
	if stats[0].FirstSeen.IsZero() {
		t.Error("FirstSeen should not be zero")
	}
	if stats[0].LastSeen.IsZero() {
		t.Error("LastSeen should not be zero")
	}
}

func TestGetSourceStats(t *testing.T) {
	dbPath := "/tmp/test_requests_source_stats.db"
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	// Log requests from multiple IPs to different endpoints
	testData := []struct {
		ip  string
		url string
	}{
		{"192.168.1.1", "/api/users"},
		{"192.168.1.1", "/api/posts"},
		{"192.168.1.1", "/health"},
		{"192.168.1.2", "/api/users"},
		{"192.168.1.2", "/api/posts"},
		{"192.168.1.3", "/health"},
	}

	for _, td := range testData {
		if err := db.LogRequest(td.ip, td.url); err != nil {
			t.Fatalf("Failed to log request: %v", err)
		}
	}

	// Get source stats
	stats, err := db.GetSourceStats()
	if err != nil {
		t.Fatalf("Failed to get source stats: %v", err)
	}

	if len(stats) != 3 {
		t.Errorf("Expected 3 sources, got %d", len(stats))
	}

	// Verify stats are ordered by count DESC
	if stats[0].IPAddress != "192.168.1.1" {
		t.Errorf("Expected 192.168.1.1 first, got %s", stats[0].IPAddress)
	}
	if stats[0].Count != 3 {
		t.Errorf("Expected count 3 for 192.168.1.1, got %d", stats[0].Count)
	}
	if stats[0].UniqueURLs != 3 {
		t.Errorf("Expected 3 unique URLs for 192.168.1.1, got %d", stats[0].UniqueURLs)
	}

	if stats[1].IPAddress != "192.168.1.2" {
		t.Errorf("Expected 192.168.1.2 second, got %s", stats[1].IPAddress)
	}
	if stats[1].Count != 2 {
		t.Errorf("Expected count 2 for 192.168.1.2, got %d", stats[1].Count)
	}
	if stats[1].UniqueURLs != 2 {
		t.Errorf("Expected 2 unique URLs for 192.168.1.2, got %d", stats[1].UniqueURLs)
	}

	// Verify timestamps are populated
	if stats[0].FirstSeen.IsZero() {
		t.Error("FirstSeen should not be zero")
	}
	if stats[0].LastSeen.IsZero() {
		t.Error("LastSeen should not be zero")
	}
}

func TestGetSummary(t *testing.T) {
	dbPath := "/tmp/test_requests_summary.db"
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	// Log requests from multiple IPs to different endpoints
	testData := []struct {
		ip  string
		url string
	}{
		{"192.168.1.1", "/api/users"},
		{"192.168.1.2", "/api/users"},
		{"192.168.1.3", "/api/posts"},
		{"192.168.1.1", "/health"},
		{"192.168.1.2", "/health"},
	}

	for _, td := range testData {
		if err := db.LogRequest(td.ip, td.url); err != nil {
			t.Fatalf("Failed to log request: %v", err)
		}
	}

	// Get summary
	summary, err := db.GetSummary()
	if err != nil {
		t.Fatalf("Failed to get summary: %v", err)
	}

	if summary.TotalRequests != 5 {
		t.Errorf("Expected 5 total requests, got %d", summary.TotalRequests)
	}
	if summary.UniqueIPs != 3 {
		t.Errorf("Expected 3 unique IPs, got %d", summary.UniqueIPs)
	}
	if summary.UniqueURLs != 3 {
		t.Errorf("Expected 3 unique URLs, got %d", summary.UniqueURLs)
	}
	if summary.FirstRequest.IsZero() {
		t.Error("FirstRequest should not be zero")
	}
	if summary.LastRequest.IsZero() {
		t.Error("LastRequest should not be zero")
	}
	if summary.FirstRequest.After(summary.LastRequest) {
		t.Error("FirstRequest should be before LastRequest")
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "database is locked",
			err:      fmt.Errorf("database is locked"),
			expected: true,
		},
		{
			name:     "database table is locked",
			err:      fmt.Errorf("database table is locked"),
			expected: true,
		},
		{
			name:     "SQLITE_BUSY",
			err:      fmt.Errorf("SQLITE_BUSY"),
			expected: true,
		},
		{
			name:     "other error",
			err:      fmt.Errorf("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryableError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestNew_InvalidPath(t *testing.T) {
	// Try to create database in non-existent directory without write permissions
	dbPath := "/nonexistent/path/test.db"

	_, err := New(dbPath)
	if err == nil {
		t.Error("Expected error creating database in invalid path")
	}
}

func TestGetEndpointStats_EmptyDatabase(t *testing.T) {
	dbPath := "/tmp/test_requests_empty_endpoint.db"
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	stats, err := db.GetEndpointStats()
	if err != nil {
		t.Fatalf("Failed to get endpoint stats: %v", err)
	}

	if len(stats) != 0 {
		t.Errorf("Expected 0 stats for empty database, got %d", len(stats))
	}
}

func TestGetSourceStats_EmptyDatabase(t *testing.T) {
	dbPath := "/tmp/test_requests_empty_source.db"
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	stats, err := db.GetSourceStats()
	if err != nil {
		t.Fatalf("Failed to get source stats: %v", err)
	}

	if len(stats) != 0 {
		t.Errorf("Expected 0 stats for empty database, got %d", len(stats))
	}
}

func TestGetSummary_EmptyDatabase(t *testing.T) {
	dbPath := "/tmp/test_requests_empty_summary.db"
	defer func() {
		if err := os.Remove(dbPath); err != nil {
			// Ignore remove errors in test cleanup
		}
	}()

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			// Ignore close errors in test cleanup
		}
	}()

	summary, err := db.GetSummary()
	if err != nil {
		t.Fatalf("Failed to get summary: %v", err)
	}

	if summary.TotalRequests != 0 {
		t.Errorf("Expected 0 total requests, got %d", summary.TotalRequests)
	}
	if summary.UniqueIPs != 0 {
		t.Errorf("Expected 0 unique IPs, got %d", summary.UniqueIPs)
	}
	if summary.UniqueURLs != 0 {
		t.Errorf("Expected 0 unique URLs, got %d", summary.UniqueURLs)
	}
}
