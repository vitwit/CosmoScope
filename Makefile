.PHONY: all build test clean lint coverage dev-deps

# Variables
BINARY_NAME=cosmoscope
MAIN_PACKAGE=./cmd/cosmoscope
GO_FILES=$(shell find . -type f -name '*.go')
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=${VERSION}"
GOLANGCI_LINT_VERSION=v1.55.2

all: lint test build

build:
	go build ${LDFLAGS} -o bin/${BINARY_NAME} ${MAIN_PACKAGE}

test:
	go test -v -race -coverprofile=coverage.out ./...

clean:
	go clean
	rm -f bin/${BINARY_NAME}
	rm -f coverage.out

lint:
	@if ! which golangci-lint >/dev/null; then \
		echo "Installing golangci-lint..." && \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}; \
	fi
	golangci-lint run

coverage: test
	go tool cover -html=coverage.out

# Install development dependencies
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

# Run the application
run: build
	./bin/${BINARY_NAME}

# Update dependencies
deps-update:
	go mod tidy
	go mod verify

# Check tools versions
check-tools:
	@echo "Checking tools versions..."
	@go version
	@golangci-lint --version