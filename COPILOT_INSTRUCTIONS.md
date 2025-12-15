# Copilot Instructions for silver-eureka

## Commit Messages
- Always use concise, descriptive commit messages
- Format: lowercase, no period, imperative mood
- Example: `improve handler coverage to 100%`

## Error Handling

### Production Code
- Never ignore errors with `_ =` - always handle them properly
- HTTP write errors: Check but don't crash (response already started)
  ```go
  if _, err := w.Write(data); err != nil {
      // Response already started
  }
  ```
- Database cleanup: Log errors but don't mask original errors
  ```go
  if closeErr := conn.Close(); closeErr != nil {
      // Log but don't mask original error
  }
  ```

### Test Code
- Use proper error handling in defer statements for cleanup:
  ```go
  defer func() {
      if err := db.Close(); err != nil {
          // Ignore close errors in test cleanup
      }
  }()
  ```
- Always check and validate errors in test assertions:
  ```go
  if err := someFunc(); err != nil {
      t.Fatalf("Expected no error, got: %v", err)
  }
  ```

## Testing
- Makefile `test` target should show total coverage
- Aim for 80%+ overall test coverage
- Target 85%+ for individual packages when improving coverage
- Add tests for error paths, not just happy paths
- Test edge cases: headers, malformed input, database errors

## Code Quality
- Follow golangci-lint rules strictly
- Use `errcheck` with `check-blank: true`
- Prefer explicit error handling over blank identifiers
- Keep code idiomatic and well-documented

## Test Coverage Priorities
1. Error paths and edge cases
2. Header parsing (X-Forwarded-For, X-Real-IP)
3. Database error conditions
4. Malformed input handling
5. Authentication/authorization failures

## CI/CD
- All changes must pass:
  - `make test` (with race detection)
  - `make lint` (golangci-lint)
  - `make build`
- Use CODECOV_TOKEN for coverage uploads

## Git Practices
- **Never use `git add -A`** - always explicitly specify files to add
- Example: `git add internal/handler/handler_test.go` instead of `git add -A`
- **Never commit without explicit user approval**
- Always show what will be committed and wait for confirmation
- **Keep commits minimal and logically grouped**
- One logical change per commit (e.g., one package's test improvements, one bug fix)
- Don't mix unrelated changes in a single commit
- Be intentional about what gets committed
