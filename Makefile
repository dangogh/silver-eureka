.PHONY: test cover clean lint

# Run all tests with race detection and generate coverage report
test:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep total:

# Open coverage report in browser
cover: test
	go tool cover -html=coverage.out

# Run linters
lint:
	golangci-lint run

# Clean up generated files
clean:
	rm -f coverage.out
	rm -f requests.db
	rm -f server.crt server.key
