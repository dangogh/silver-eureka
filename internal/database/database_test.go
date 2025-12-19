package database

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) *DB {
	t.Helper()
	dbPath := fmt.Sprintf("/tmp/test_db_%d.db", time.Now().UnixNano())
	t.Cleanup(func() {
		_ = os.Remove(dbPath) //nolint:errcheck // Cleanup code
	})

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close() //nolint:errcheck // Cleanup code
	})

	return db
}

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

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "normal input",
			input:  "192.168.1.1",
			maxLen: 50,
			want:   "192.168.1.1",
		},
		{
			name:   "input with newline",
			input:  "test\nvalue",
			maxLen: 50,
			want:   "testvalue",
		},
		{
			name:   "input with carriage return",
			input:  "test\rvalue",
			maxLen: 50,
			want:   "testvalue",
		},
		{
			name:   "input with null byte",
			input:  "test\x00value",
			maxLen: 50,
			want:   "testvalue",
		},
		{
			name:   "input with tab",
			input:  "test\tvalue",
			maxLen: 50,
			want:   "testvalue",
		},
		{
			name:   "input exceeds max length",
			input:  "very long string that exceeds the maximum allowed length",
			maxLen: 10,
			want:   "very long ",
		},
		{
			name:   "URL with control chars",
			input:  "/api/test\n\r\x00?param=value",
			maxLen: 100,
			want:   "/api/test?param=value",
		},
		{
			name:   "empty input",
			input:  "",
			maxLen: 50,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeInput(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("sanitizeInput(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestLogRequest_WithControlCharacters(t *testing.T) {
	dbPath := "/tmp/test_requests_sanitize.db"
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

	// Log request with control characters
	maliciousURL := "/test\n\r\x00path"
	maliciousIP := "192.168\n.1.1"

	if err := db.LogRequest(maliciousIP, maliciousURL); err != nil {
		t.Errorf("LogRequest failed: %v", err)
	}

	// Verify the log was sanitized
	logs, err := db.GetLogs(1)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	// Should have control characters stripped
	if strings.Contains(logs[0].IPAddress, "\n") || strings.Contains(logs[0].IPAddress, "\r") {
		t.Errorf("IP address still contains control characters: %q", logs[0].IPAddress)
	}

	if strings.Contains(logs[0].URL, "\n") || strings.Contains(logs[0].URL, "\r") || strings.Contains(logs[0].URL, "\x00") {
		t.Errorf("URL still contains control characters: %q", logs[0].URL)
	}

	// Verify sanitized values
	expectedIP := "192.168.1.1"
	expectedURL := "/testpath"

	if logs[0].IPAddress != expectedIP {
		t.Errorf("IP address = %q, want %q", logs[0].IPAddress, expectedIP)
	}

	if logs[0].URL != expectedURL {
		t.Errorf("URL = %q, want %q", logs[0].URL, expectedURL)
	}
}

func TestLogRequest_LongInputs(t *testing.T) {
	dbPath := "/tmp/test_requests_long.db"
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

	// Test with very long IP (should be truncated to 45 chars)
	longIP := strings.Repeat("1234567890", 10) // 100 chars
	normalURL := "/test"

	if err := db.LogRequest(longIP, normalURL); err != nil {
		t.Errorf("LogRequest failed: %v", err)
	}

	// Test with very long URL (should be truncated to 2048 chars)
	normalIP := "192.168.1.1"
	longURL := "/" + strings.Repeat("abcdefghij", 300) // >2048 chars

	if err := db.LogRequest(normalIP, longURL); err != nil {
		t.Errorf("LogRequest failed: %v", err)
	}

	// Verify truncation occurred
	logs, err := db.GetLogs(10)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 2 {
		t.Fatalf("Expected 2 log entries, got %d", len(logs))
	}

	// Check IP was truncated
	for _, log := range logs {
		if len(log.IPAddress) > 45 {
			t.Errorf("IP address too long: %d chars (max 45)", len(log.IPAddress))
		}
		if len(log.URL) > 2048 {
			t.Errorf("URL too long: %d chars (max 2048)", len(log.URL))
		}
	}
}

func TestLogRequest_EmptyInputs(t *testing.T) {
	dbPath := "/tmp/test_requests_empty.db"
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

	// Test with empty strings
	if err := db.LogRequest("", ""); err != nil {
		t.Errorf("LogRequest with empty inputs failed: %v", err)
	}

	// Verify the log was created
	logs, err := db.GetLogs(1)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	if logs[0].IPAddress != "" {
		t.Errorf("Expected empty IP, got %q", logs[0].IPAddress)
	}

	if logs[0].URL != "" {
		t.Errorf("Expected empty URL, got %q", logs[0].URL)
	}
}

func TestExecuteWithRetry_Success(t *testing.T) {
	db := setupTestDB(t)

	// Operation succeeds on first try
	callCount := 0
	err := db.executeWithRetry(func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected operation to be called once, got: %d", callCount)
	}
}

func TestExecuteWithRetry_NonRetryableError(t *testing.T) {
	db := setupTestDB(t)

	// Non-retryable error should fail immediately
	callCount := 0
	expectedErr := errors.New("non-retryable error")
	err := db.executeWithRetry(func() error {
		callCount++
		return expectedErr
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to execute operation") {
		t.Fatalf("expected 'failed to execute operation' error, got: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected operation to be called once for non-retryable error, got: %d", callCount)
	}
}

func TestExecuteWithRetry_SuccessAfterRetries(t *testing.T) {
	db := setupTestDB(t)

	// Operation succeeds on third try (simulate database busy)
	callCount := 0
	err := db.executeWithRetry(func() error {
		callCount++
		if callCount < 3 {
			return sqlite3.Error{Code: sqlite3.ErrBusy}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error after retries, got: %v", err)
	}
	if callCount != 3 {
		t.Fatalf("expected operation to be called 3 times, got: %d", callCount)
	}
}

func TestExecuteWithRetry_ExhaustedRetries(t *testing.T) {
	db := setupTestDB(t)

	// Operation fails all attempts
	callCount := 0
	err := db.executeWithRetry(func() error {
		callCount++
		return sqlite3.Error{Code: sqlite3.ErrBusy}
	})

	if err == nil {
		t.Fatal("expected error after exhausted retries, got nil")
	}
	if !strings.Contains(err.Error(), "failed to execute operation after") {
		t.Fatalf("expected 'failed after retries' error, got: %v", err)
	}
	// Should be called 4 times total (initial + 3 retries)
	if callCount != 4 {
		t.Fatalf("expected operation to be called 4 times (initial + 3 retries), got: %d", callCount)
	}
}

func TestCleanupOldLogs_NoRetention(t *testing.T) {
	db := setupTestDB(t)

	// With retention = 0, no cleanup should occur
	deleted, err := db.CleanupOldLogs(0)
	if err != nil {
		t.Fatalf("expected no error with retention=0, got: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted with retention=0, got: %d", deleted)
	}

	// Negative retention should also do nothing
	deleted, err = db.CleanupOldLogs(-1)
	if err != nil {
		t.Fatalf("expected no error with retention=-1, got: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted with retention=-1, got: %d", deleted)
	}
}

func TestCleanupOldLogs_DeletesOldRecords(t *testing.T) {
	db := setupTestDB(t)

	// Insert logs with different timestamps
	now := time.Now()

	// Old logs (40 days ago)
	oldTime := now.AddDate(0, 0, -40)
	_, err := db.conn.Exec(
		"INSERT INTO request_logs (ip_address, url, timestamp) VALUES (?, ?, ?)",
		"192.0.2.1", "/old1", oldTime,
	)
	if err != nil {
		t.Fatalf("Failed to insert old log: %v", err)
	}

	_, err = db.conn.Exec(
		"INSERT INTO request_logs (ip_address, url, timestamp) VALUES (?, ?, ?)",
		"192.0.2.2", "/old2", oldTime.Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("Failed to insert old log: %v", err)
	}

	// Recent logs (10 days ago)
	recentTime := now.AddDate(0, 0, -10)
	_, err = db.conn.Exec(
		"INSERT INTO request_logs (ip_address, url, timestamp) VALUES (?, ?, ?)",
		"192.0.2.3", "/recent1", recentTime,
	)
	if err != nil {
		t.Fatalf("Failed to insert recent log: %v", err)
	}

	// Very recent (1 day ago)
	veryRecentTime := now.AddDate(0, 0, -1)
	_, err = db.conn.Exec(
		"INSERT INTO request_logs (ip_address, url, timestamp) VALUES (?, ?, ?)",
		"192.0.2.4", "/recent2", veryRecentTime,
	)
	if err != nil {
		t.Fatalf("Failed to insert very recent log: %v", err)
	}

	// Cleanup logs older than 30 days
	deleted, err := db.CleanupOldLogs(30)
	if err != nil {
		t.Fatalf("Failed to cleanup old logs: %v", err)
	}

	// Should have deleted 2 old logs
	if deleted != 2 {
		t.Errorf("Expected 2 deleted logs, got: %d", deleted)
	}

	// Verify remaining logs
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM request_logs").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count remaining logs: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 remaining logs, got: %d", count)
	}

	// Verify the correct logs remain (recent ones)
	logs, err := db.GetLogs(10)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 2 {
		t.Fatalf("Expected 2 logs, got: %d", len(logs))
	}

	// Check that recent logs are present (order is DESC by timestamp)
	if logs[0].URL != "/recent2" {
		t.Errorf("Expected first log URL to be /recent2, got: %s", logs[0].URL)
	}
	if logs[1].URL != "/recent1" {
		t.Errorf("Expected second log URL to be /recent1, got: %s", logs[1].URL)
	}
}

func TestCleanupOldLogs_NoOldRecords(t *testing.T) {
	db := setupTestDB(t)

	// Insert only recent logs
	now := time.Now()
	recentTime := now.AddDate(0, 0, -5)

	if err := db.LogRequest("192.0.2.1", "/test1"); err != nil {
		t.Fatalf("Failed to log request: %v", err)
	}

	// Manually update timestamp to be 5 days old
	_, err := db.conn.Exec(
		"UPDATE request_logs SET timestamp = ?",
		recentTime,
	)
	if err != nil {
		t.Fatalf("Failed to update timestamp: %v", err)
	}

	// Cleanup with 30 day retention
	deleted, err := db.CleanupOldLogs(30)
	if err != nil {
		t.Fatalf("Failed to cleanup: %v", err)
	}

	// Should have deleted 0 records
	if deleted != 0 {
		t.Errorf("Expected 0 deleted logs (all are recent), got: %d", deleted)
	}

	// Verify log still exists
	logs, err := db.GetLogs(10)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 remaining log, got: %d", len(logs))
	}
}

func TestCleanupOldLogs_AllOldRecords(t *testing.T) {
	db := setupTestDB(t)

	// Insert only old logs
	now := time.Now()
	oldTime := now.AddDate(0, 0, -60)

	for i := 0; i < 5; i++ {
		_, err := db.conn.Exec(
			"INSERT INTO request_logs (ip_address, url, timestamp) VALUES (?, ?, ?)",
			"192.0.2.1", fmt.Sprintf("/old%d", i), oldTime.Add(time.Duration(i)*time.Hour),
		)
		if err != nil {
			t.Fatalf("Failed to insert old log: %v", err)
		}
	}

	// Cleanup with 30 day retention
	deleted, err := db.CleanupOldLogs(30)
	if err != nil {
		t.Fatalf("Failed to cleanup: %v", err)
	}

	// Should have deleted all 5 records
	if deleted != 5 {
		t.Errorf("Expected 5 deleted logs, got: %d", deleted)
	}

	// Verify no logs remain
	logs, err := db.GetLogs(10)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 0 {
		t.Errorf("Expected 0 remaining logs, got: %d", len(logs))
	}
}
