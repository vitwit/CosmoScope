BINARY_NAME=cosmoscope
MAIN_PACKAGE=./cmd/cosmoscope
GO_FILES=$(shell find . -type f -name '*.go')
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=${VERSION}"

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
	golangci-lint run

coverage: test
	go tool cover -html=coverage.out

# Install development dependencies
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run the application
run: build
	./bin/${BINARY_NAME}

# Update dependencies
deps-update:
	go mod tidy
	go mod verify

.PHONY: all build test clean lint coverage
