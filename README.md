# silver-eureka

A secure HTTPS-only Go web application that logs requests to an SQLite database.

## Features

- **HTTPS-only server** using Go's standard library
- TLS with configurable certificates (required)
- Default HTTPS port 443 with override via `--port` flag
- Structured JSON logging with debug level for request details
- Logs all HTTPS requests with IP address and URL to SQLite database
- **Statistics endpoints** for analyzing logged requests:
  - Overall summary statistics
  - Statistics grouped by endpoint/URL
  - Statistics grouped by source IP address
- Graceful shutdown handling
- Comprehensive test coverage

## Requirements

- Go 1.24 or later

## Installation

```bash
go build -o app .
```

## Usage

### Running the Server

#### Default HTTPS (port 443)
```bash
# Generate self-signed certificate for testing
./generate-cert.sh localhost

# Run with default HTTPS port 443
sudo ./app

# Or run on a custom port (no sudo required for ports > 1024)
./app -port=8443
```

#### Custom Port
```bash
./app -port=8443
```

### Configuration

The server is HTTPS-only and requires TLS certificates. The default certificate paths are `server.crt` and `server.key` in the current directory.

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `443` | HTTPS server port |

**Note**: Port 443 requires sudo/root privileges. For development, use a port above 1024 (e.g., 8443).

### Testing

Run all tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

Using Make targets:
```bash
# Run tests with race detection and generate coverage report
make test

# View coverage report in browser
make cover

# Clean up generated files (coverage, database, certificates)
make clean
```

### API

The application provides both request logging and statistics endpoints.

#### Request Logging

Any request to paths other than `/stats/*` will be logged to the database with:
- Client IP address (supports X-Forwarded-For and X-Real-IP headers)
- Requested URL path
- Timestamp

All requests are logged in JSON format with debug-level details including headers, user agent, and more.

Example request (use -k for self-signed cert):
```bash
curl -k https://localhost/any/path
```

Response:
```
Request logged: /any/path from 127.0.0.1
```

#### Statistics Endpoints

The application provides three statistics endpoints that return JSON data:

**GET /stats/summary** - Overall statistics
```bash
curl -k https://localhost/stats/summary
```
Response:
```json
{
  "total_requests": 150,
  "unique_ips": 23,
  "unique_urls": 45,
  "first_request": "2025-12-06T10:00:00Z",
  "last_request": "2025-12-06T17:30:00Z"
}
```

**GET /stats/endpoints** - Statistics grouped by endpoint/URL
```bash
curl -k https://localhost/stats/endpoints
```
Response:
```json
[
  {
    "url": "/api/users",
    "count": 45,
    "first_seen": "2025-12-06T10:00:00Z",
    "last_seen": "2025-12-06T17:30:00Z",
    "unique_ips": 12
  },
  {
    "url": "/api/products",
    "count": 30,
    "first_seen": "2025-12-06T10:15:00Z",
    "last_seen": "2025-12-06T17:25:00Z",
    "unique_ips": 8
  }
]
```

**GET /stats/sources** - Statistics grouped by IP address/source
```bash
curl -k https://localhost/stats/sources
```
Response:
```json
[
  {
    "ip_address": "192.168.1.100",
    "count": 25,
    "first_seen": "2025-12-06T10:00:00Z",
    "last_seen": "2025-12-06T17:30:00Z",
    "unique_urls": 10
  },
  {
    "ip_address": "192.168.1.101",
    "count": 18,
    "first_seen": "2025-12-06T10:30:00Z",
    "last_seen": "2025-12-06T17:20:00Z",
    "unique_urls": 7
  }
]
```

## Database

The application uses an SQLite database file named `requests.db` to store all request logs. The database is automatically created on first run with the following schema:

```sql
CREATE TABLE request_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip_address TEXT NOT NULL,
    url TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Project Structure

```
.
├── main.go                      # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   ├── database/                # SQLite database operations
│   │   ├── database.go
│   │   └── database_test.go
│   └── handler/                 # HTTP request handlers
│       ├── handler.go
│       └── handler_test.go
├── go.mod
├── go.sum
└── README.md
```

## License

See [LICENSE](LICENSE) file for details.