package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the sql.DB connection
type DB struct {
	conn *sql.DB
}

// RequestLog represents a logged HTTP request
type RequestLog struct {
	ID        int64
	IPAddress string
	URL       string
	Timestamp time.Time
}

// EndpointStats represents statistics for a specific endpoint
type EndpointStats struct {
	URL       string    `json:"url"`
	Count     int64     `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	UniqueIPs int64     `json:"unique_ips"`
}

// SourceStats represents statistics for a specific IP address
type SourceStats struct {
	IPAddress  string    `json:"ip_address"`
	Count      int64     `json:"count"`
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
	UniqueURLs int64     `json:"unique_urls"`
}

// Summary represents overall statistics
type Summary struct {
	TotalRequests int64     `json:"total_requests"`
	UniqueIPs     int64     `json:"unique_ips"`
	UniqueURLs    int64     `json:"unique_urls"`
	FirstRequest  time.Time `json:"first_request"`
	LastRequest   time.Time `json:"last_request"`
}

// New creates a new database connection and initializes the schema
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize schema
	if err := db.initSchema(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// initSchema creates the necessary tables if they don't exist
func (db *DB) initSchema() error {
	// Configure SQLite for better performance
	pragmas := `
	PRAGMA journal_mode = WAL;
	PRAGMA synchronous = NORMAL;
	PRAGMA cache_size = -64000;
	PRAGMA busy_timeout = 5000;
	`
	if _, err := db.conn.Exec(pragmas); err != nil {
		return fmt.Errorf("failed to set pragmas: %w", err)
	}

	// Set connection pool limits
	db.conn.SetMaxOpenConns(1) // SQLite only supports one writer
	db.conn.SetMaxIdleConns(1)
	db.conn.SetConnMaxLifetime(0)

	query := `
	CREATE TABLE IF NOT EXISTS request_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip_address TEXT NOT NULL,
		url TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON request_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_ip_address ON request_logs(ip_address);
	CREATE INDEX IF NOT EXISTS idx_url ON request_logs(url);
	`

	_, err := db.conn.Exec(query)
	return err
}

// LogRequest logs an HTTP request to the database
func (db *DB) LogRequest(ipAddress, url string) error {
	query := `INSERT INTO request_logs (ip_address, url, timestamp) VALUES (?, ?, ?)`
	_, err := db.conn.Exec(query, ipAddress, url, time.Now())
	if err != nil {
		return fmt.Errorf("failed to log request: %w", err)
	}
	return nil
}

// GetLogs retrieves request logs with optional limit
func (db *DB) GetLogs(limit int) ([]RequestLog, error) {
	var query string
	var rows *sql.Rows
	var err error

	if limit > 0 {
		query = `SELECT id, ip_address, url, timestamp FROM request_logs ORDER BY timestamp DESC LIMIT ?`
		rows, err = db.conn.Query(query, limit)
	} else {
		query = `SELECT id, ip_address, url, timestamp FROM request_logs ORDER BY timestamp DESC`
		rows, err = db.conn.Query(query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	var logs []RequestLog
	for rows.Next() {
		var log RequestLog
		if err := rows.Scan(&log.ID, &log.IPAddress, &log.URL, &log.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return logs, nil
}

// GetAllLogs retrieves all request logs from the database with a safety limit
func (db *DB) GetAllLogs() ([]RequestLog, error) {
	// Limit to 100k records to prevent memory exhaustion
	// For larger exports, implement pagination or streaming
	return db.GetLogs(100000)
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// GetEndpointStats retrieves statistics grouped by endpoint/URL
func (db *DB) GetEndpointStats() ([]EndpointStats, error) {
	query := `
		SELECT 
			url,
			COUNT(*) as count,
			MIN(timestamp) as first_seen,
			MAX(timestamp) as last_seen,
			COUNT(DISTINCT ip_address) as unique_ips
		FROM request_logs
		GROUP BY url
		ORDER BY count DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query endpoint stats: %w", err)
	}
	defer rows.Close()

	var stats []EndpointStats
	for rows.Next() {
		var s EndpointStats
		var firstSeen, lastSeen string
		if err := rows.Scan(&s.URL, &s.Count, &firstSeen, &lastSeen, &s.UniqueIPs); err != nil {
			return nil, fmt.Errorf("failed to scan endpoint stats: %w", err)
		}
		// Parse timestamps
		if s.FirstSeen, err = time.Parse("2006-01-02 15:04:05.999999999-07:00", firstSeen); err != nil {
			if s.FirstSeen, err = time.Parse("2006-01-02 15:04:05", firstSeen); err != nil {
				return nil, fmt.Errorf("failed to parse first_seen: %w", err)
			}
		}
		if s.LastSeen, err = time.Parse("2006-01-02 15:04:05.999999999-07:00", lastSeen); err != nil {
			if s.LastSeen, err = time.Parse("2006-01-02 15:04:05", lastSeen); err != nil {
				return nil, fmt.Errorf("failed to parse last_seen: %w", err)
			}
		}
		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("endpoint stats iteration error: %w", err)
	}

	return stats, nil
}

// GetSourceStats retrieves statistics grouped by IP address
func (db *DB) GetSourceStats() ([]SourceStats, error) {
	query := `
		SELECT 
			ip_address,
			COUNT(*) as count,
			MIN(timestamp) as first_seen,
			MAX(timestamp) as last_seen,
			COUNT(DISTINCT url) as unique_urls
		FROM request_logs
		GROUP BY ip_address
		ORDER BY count DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query source stats: %w", err)
	}
	defer rows.Close()

	var stats []SourceStats
	for rows.Next() {
		var s SourceStats
		var firstSeen, lastSeen string
		if err := rows.Scan(&s.IPAddress, &s.Count, &firstSeen, &lastSeen, &s.UniqueURLs); err != nil {
			return nil, fmt.Errorf("failed to scan source stats: %w", err)
		}
		// Parse timestamps
		if s.FirstSeen, err = time.Parse("2006-01-02 15:04:05.999999999-07:00", firstSeen); err != nil {
			if s.FirstSeen, err = time.Parse("2006-01-02 15:04:05", firstSeen); err != nil {
				return nil, fmt.Errorf("failed to parse first_seen: %w", err)
			}
		}
		if s.LastSeen, err = time.Parse("2006-01-02 15:04:05.999999999-07:00", lastSeen); err != nil {
			if s.LastSeen, err = time.Parse("2006-01-02 15:04:05", lastSeen); err != nil {
				return nil, fmt.Errorf("failed to parse last_seen: %w", err)
			}
		}
		stats = append(stats, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("source stats iteration error: %w", err)
	}

	return stats, nil
}

// GetSummary retrieves overall statistics
func (db *DB) GetSummary() (*Summary, error) {
	query := `
		SELECT 
			COUNT(*) as total_requests,
			COUNT(DISTINCT ip_address) as unique_ips,
			COUNT(DISTINCT url) as unique_urls,
			MIN(timestamp) as first_request,
			MAX(timestamp) as last_request
		FROM request_logs
	`

	var summary Summary
	var firstRequest, lastRequest string
	err := db.conn.QueryRow(query).Scan(
		&summary.TotalRequests,
		&summary.UniqueIPs,
		&summary.UniqueURLs,
		&firstRequest,
		&lastRequest,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query summary stats: %w", err)
	}

	// Parse timestamps
	if summary.FirstRequest, err = time.Parse("2006-01-02 15:04:05.999999999-07:00", firstRequest); err != nil {
		if summary.FirstRequest, err = time.Parse("2006-01-02 15:04:05", firstRequest); err != nil {
			return nil, fmt.Errorf("failed to parse first_request: %w", err)
		}
	}
	if summary.LastRequest, err = time.Parse("2006-01-02 15:04:05.999999999-07:00", lastRequest); err != nil {
		if summary.LastRequest, err = time.Parse("2006-01-02 15:04:05", lastRequest); err != nil {
			return nil, fmt.Errorf("failed to parse last_request: %w", err)
		}
	}

	return &summary, nil
}
