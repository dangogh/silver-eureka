# silver-eureka

A Go web application that logs HTTP requests to an SQLite database with optional HTTP Basic Authentication.

## Features

- **HTTP server** on port 8080 (configurable)
- **Optional HTTP Basic Authentication** to protect statistics endpoints
- **Web interface** with session-based authentication for easy stats viewing
- Structured JSON logging with debug level for request details
- Logs all HTTP requests with IP address and URL to SQLite database
- **Statistics endpoints** for analyzing logged requests:
  - Overall summary statistics
  - Statistics grouped by endpoint/URL
  - Statistics grouped by source IP address
  - Downloadable CSV export
- **Health check endpoint** for monitoring
- Graceful shutdown handling
- Docker support with health checks
- Comprehensive test coverage

## Requirements

- Go 1.24 or later

## Installation

```bash
go build -o app .
```

## Usage

### Running the Server

#### Default HTTP (port 8080)
```bash
# Without authentication (stats endpoints are public)
./app

# With authentication (protects /stats/* endpoints)
AUTH_USERNAME=admin AUTH_PASSWORD=secret123 ./app

# Custom port
./app -port=9000

# Custom database path
./app -db=/path/to/requests.db
```

#### Using Docker
```bash
# Build and run with Docker Compose
docker-compose up -d

# Enable authentication by editing docker-compose.yml
# Uncomment and set AUTH_USERNAME and AUTH_PASSWORD
```

### Configuration

| Environment Variable | Flag | Default | Description |
|---------------------|------|---------|-------------|
| `PORT` | `-port` | `8080` | HTTP server port |
| `DB_PATH` | `-db` | `data/requests.db` | SQLite database file path |
| `AUTH_USERNAME` | `-auth-user` | `""` | Username for HTTP Basic Auth (optional) |
| `AUTH_PASSWORD` | `-auth-pass` | `""` | Password for HTTP Basic Auth (optional) |

**Authentication**: When `AUTH_USERNAME` and `AUTH_PASSWORD` are set, all `/stats/*` endpoints require HTTP Basic Authentication. The `/health` and logging endpoints remain public.

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

The application provides both a web interface and API endpoints.

#### Web Interface

When authentication is configured, access the web interface at:

```
http://localhost:8080/login
```

After logging in, you'll see a dashboard with links to view all statistics in formatted HTML tables.

**Features:**
- Session-based authentication (24-hour timeout)
- Dashboard with stat cards
- Formatted HTML views for all statistics
- Logout functionality

#### Request Logging

Any request to paths other than `/stats/*` will be logged to the database with:
- Client IP address (supports X-Forwarded-For and X-Real-IP headers)
- Requested URL path
- Timestamp

All requests are logged in JSON format with debug-level details including headers, user agent, and more.

Example request:
```bash
curl http://localhost:8080/any/path
```

Response:
```
Request logged: /any/path from 127.0.0.1
```

#### Health Check

The `/health` endpoint provides service health status (always public, no auth required):

```bash
curl http://localhost:8080/health
```

Response:
```json
{"status":"healthy","database":"up"}
```

#### Statistics Endpoints

**Note**: When authentication is enabled via `AUTH_USERNAME` and `AUTH_PASSWORD`, these endpoints require HTTP Basic Auth credentials.

**GET /stats/summary** - Overall statistics
```bash
# Without auth
curl http://localhost:8080/stats/summary

# With auth
curl -u admin:secret123 http://localhost:8080/stats/summary
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
curl -u admin:secret123 http://localhost:8080/stats/endpoints
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
curl -u admin:secret123 http://localhost:8080/stats/sources
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