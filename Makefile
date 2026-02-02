.PHONY: build test clean install run-init lint coverage quality

# Build the binary
build:
	go build -o bin/bagboy ./cmd/bagboy

# Build for all platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o dist/bagboy-darwin-amd64 ./cmd/bagboy
	GOOS=darwin GOARCH=arm64 go build -o dist/bagboy-darwin-arm64 ./cmd/bagboy
	GOOS=linux GOARCH=amd64 go build -o dist/bagboy-linux-amd64 ./cmd/bagboy
	GOOS=linux GOARCH=arm64 go build -o dist/bagboy-linux-arm64 ./cmd/bagboy
	GOOS=windows GOARCH=amd64 go build -o dist/bagboy-windows-amd64.exe ./cmd/bagboy

# Test the application
test:
	go test -v -race ./...

# Test with coverage
coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...
	goimports -w .

# Vet the code
vet:
	go vet ./...

# Run all quality checks
quality: fmt vet lint test coverage
	@echo "✅ All quality checks passed"

# Clean build artifacts
clean:
	rm -rf bin/ dist/ coverage.out coverage.html

# Install locally
install: build
	cp bin/bagboy /usr/local/bin/

# Run init command for testing
run-init: build
	./bin/bagboy init --interactive

# Run pack command for testing
run-pack: build
	./bin/bagboy pack --all

# Validate config
validate: build
	./bin/bagboy validate

# Install development dependencies
deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Run security scan
security:
	gosec ./...

# Generate mocks (if needed)
mocks:
	go generate ./...
