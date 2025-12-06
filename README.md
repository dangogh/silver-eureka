# silver-eureka

A Go web application that logs HTTP requests to an SQLite database.

## Features

- HTTP server using Go's standard library
- Configurable port via environment variable or command-line flag
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

#### Default port (8080)
```bash
./app
```

#### Using environment variable
```bash
PORT=9090 ./app
```

#### Using command-line flag (highest precedence)
```bash
./app -port=7070
```

### Configuration Priority

The port configuration follows this precedence order (highest to lowest):
1. Command-line flag (`-port`)
2. Environment variable (`PORT`)
3. Default value (`8080`)

### Testing

Run all tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

### API

The application accepts requests to any path. Each request is logged to the SQLite database with:
- Client IP address (supports X-Forwarded-For and X-Real-IP headers)
- Requested URL path
- Timestamp

Example request:
```bash
curl http://localhost:8080/any/path
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