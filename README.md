# silver-eureka

A Go web application that logs HTTP requests to an SQLite database.

## Features

- HTTP/HTTPS server using Go's standard library
- TLS support with configurable certificates
- Automatic HTTP to HTTPS redirect when TLS is enabled
- Configurable ports via environment variable or command-line flag
- Structured JSON logging with debug level for request details
- Logs all HTTP requests with IP address and URL to SQLite database
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

The application accepts requests to any path. Each request is logged to the SQLite database with:
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