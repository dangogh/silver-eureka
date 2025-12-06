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
	query := `
	CREATE TABLE IF NOT EXISTS request_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip_address TEXT NOT NULL,
		url TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON request_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_ip_address ON request_logs(ip_address);
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

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}
