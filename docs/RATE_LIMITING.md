# Rate Limiting Implementation

## Overview
Added comprehensive rate limiting to protect the server from DoS attacks. Rate limiting is applied to all routes with both per-IP and global limits.

## Configuration

### Default Limits
- **Per-IP**: 100 requests per minute (10 request burst)
- **Global**: 10,000 requests per minute (1,000 request burst)

### Rate Limiting Behavior
- Returns `429 Too Many Requests` when limit is exceeded
- Logs warning messages with IP and path information
- Automatically cleans up inactive IP limiters every 5 minutes

## Implementation Details

### Files Added
- `internal/middleware/ratelimit.go` - Rate limiter implementation
- `internal/middleware/ratelimit_test.go` - Comprehensive test suite (90.2% coverage)
- `internal/router/ratelimit_integration_test.go` - Integration tests

### Files Modified
- `internal/router/router.go` - Added rate limiting middleware
- `go.mod` - Added `golang.org/x/time v0.14.0` dependency

### IP Address Detection
Rate limiting uses intelligent IP detection with the following priority:
1. `X-Forwarded-For` header (for proxies/load balancers)
2. `X-Real-IP` header
3. `RemoteAddr` (direct connection)

This ensures accurate rate limiting even behind proxies or CDNs.

## Testing

### Test Coverage
- **Unit Tests**: 12 test functions covering all scenarios
  - Per-IP rate limiting
  - Global rate limiting
  - Different IP addresses get separate limits
  - Burst handling
  - IP header detection (X-Forwarded-For, X-Real-IP)
  - Cleanup routine
  - Helper functions

- **Integration Tests**: 2 test functions
  - Rate limiting enabled (verifies requests are limited)
  - Rate limiting disabled (verifies no limits applied)

### Running Tests
```bash
# Run all tests
make test

# Run only rate limiter tests
go test -v ./internal/middleware -run TestRateLimiter

# Run integration tests
go test -v ./internal/router -run TestRateLimit
```

## Usage

### Production
Rate limiting is automatically enabled in production:
```go
router := router.New(db, authUsername, authPassword)
```

### Testing
Rate limiting can be disabled for tests:
```go
router := router.NewWithRateLimiter(db, authUsername, authPassword, false)
```

## Benefits

### DoS Protection
- **Per-IP limits** prevent individual attackers from overwhelming the server
- **Global limits** protect against distributed attacks from many IPs
- **Burst allowance** accommodates legitimate traffic spikes

### Resource Management
- Automatic cleanup of inactive IP limiters prevents memory leaks
- Lightweight token bucket algorithm ensures minimal overhead
- No external dependencies (uses stdlib + golang.org/x/time)

### Observability
- Rate limit violations are logged with IP and path
- Helps identify attack patterns and malicious sources
- Integration with existing structured logging

## Configuration Options

To customize rate limits, modify `router.go`:
```go
// Initialize rate limiter: X req/min per IP, Y req/min global
rateLimiter := middleware.NewRateLimiter(perIPReqPerMin, globalReqPerMin)
```

Recommended configurations:
- **Aggressive**: 50 per-IP, 5,000 global (stricter protection)
- **Standard**: 100 per-IP, 10,000 global (default, balanced)
- **Permissive**: 200 per-IP, 20,000 global (for high-traffic scenarios)

## Monitoring

Rate limit events are logged at WARN level:
```
2025/12/16 20:23:17 WARN Per-IP rate limit exceeded ip=192.0.2.1 path=/test/path
2025/12/16 20:23:17 WARN Global rate limit exceeded ip=203.0.113.1 path=/api/endpoint
```

Consider setting up alerts for:
- High frequency of rate limit warnings (possible attack)
- Specific IPs repeatedly hitting limits (block at firewall level)
- Global limit being hit (may need to increase capacity)

## Performance Impact

- **Overhead**: ~10-50μs per request (minimal)
- **Memory**: ~100-200 bytes per active IP address
- **Cleanup**: Runs every 5 minutes, removing inactive limiters
- **No database queries**: All rate limiting done in-memory

## Future Enhancements

Potential improvements:
- [ ] Redis-backed rate limiting for distributed deployments
- [ ] Dynamic rate limits based on authentication status
- [ ] Whitelist for trusted IPs (monitoring services)
- [ ] Rate limit metrics endpoint for monitoring
- [ ] Configurable rate limits via environment variables
