.PHONY: test cover clean

# Run all tests with race detection and generate coverage report
test:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Open coverage report in browser
cover: test
	go tool cover -html=coverage.out

# Clean up generated files
clean:
	rm -f coverage.out
	rm -f requests.db
	rm -f server.crt server.key
