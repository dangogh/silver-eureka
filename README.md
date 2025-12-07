# silver-eureka

A Go web application that logs HTTP requests to an SQLite database.

## Features

- HTTP/HTTPS server using Go's standard library
- TLS support with configurable certificates
- Automatic HTTP to HTTPS redirect when TLS is enabled
- Configurable ports via environment variable or command-line flag
- Structured JSON logging with debug level for request details
- Logs all HTTP requests with IP address and URL to SQLite database
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

#### Default port (8080) - HTTP
```bash
./app
```

#### With TLS enabled (HTTPS)
```bash
# Generate self-signed certificate for testing
./generate-cert.sh localhost

# Run with TLS (automatically starts HTTP redirect server on port 8000)
./app -tls -tls-cert=server.crt -tls-key=server.key

# Run with TLS but disable HTTP redirect
./app -tls -tls-cert=server.crt -tls-key=server.key -http-redirect=false

# Run with TLS on custom ports
./app -tls -port=8443 -http-port=8080 -tls-cert=server.crt -tls-key=server.key
```

#### Using environment variables
```bash
# HTTP on custom port
PORT=9090 ./app

# HTTPS with custom certificates
export TLS_ENABLED=true
export TLS_CERT=/path/to/server.crt
export TLS_KEY=/path/to/server.key
./app
```

#### Using command-line flags (highest precedence)
```bash
./app -port=7070 -tls -tls-cert=server.crt -tls-key=server.key
```

### Configuration Priority

Configuration follows this precedence order (highest to lowest):
1. Command-line flags (`-port`, `-http-port`, `-tls`, `-tls-cert`, `-tls-key`, `-http-redirect`)
2. Environment variables (`PORT`, `HTTP_PORT`, `TLS_ENABLED`, `TLS_CERT`, `TLS_KEY`, `HTTP_REDIRECT`)
3. Default values

### HTTP to HTTPS Redirect

When TLS is enabled, the application automatically starts two servers:
- **HTTPS server** on the configured port (default: 8080)
- **HTTP redirect server** on the HTTP port (default: 8000)

The HTTP server automatically redirects all requests to HTTPS with a 301 (Moved Permanently) status. Each redirect is logged in JSON format with debug and info level messages.

To disable the automatic redirect and only run the HTTPS server:
```bash
./app -tls -http-redirect=false
```

### Configuration Options

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `-port` | `PORT` | `8080` | HTTPS server port (or HTTP if TLS disabled) |
| `-http-port` | `HTTP_PORT` | `8000` | HTTP redirect server port (when TLS enabled) |
| `-tls` | `TLS_ENABLED` | `false` | Enable TLS/HTTPS |
| `-tls-cert` | `TLS_CERT` | `server.crt` | Path to TLS certificate file |
| `-tls-key` | `TLS_KEY` | `server.key` | Path to TLS private key file |
| `-http-redirect` | `HTTP_REDIRECT` | `true` | Enable HTTP to HTTPS redirect when TLS is enabled |

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

Example HTTP request:
```bash
curl http://localhost:8080/any/path
```

Example HTTPS request (with self-signed cert):
```bash
curl -k https://localhost:8080/any/path
```

Response:
```
Request logged: /any/path from 127.0.0.1
```

#### Statistics Endpoints

The application provides three statistics endpoints that return JSON data:

**GET /stats/summary** - Overall statistics
```bash
curl http://localhost:8080/stats/summary
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
curl http://localhost:8080/stats/endpoints
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
curl http://localhost:8080/stats/sources
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