.PHONY: run build test vet lint fmt tidy proto

BINARY := bin/server

## Start the server locally (loads .env from repo root)
run:
	go run ./cmd/server

## Build the server binary
build:
	go build -o $(BINARY) ./cmd/server

## Run all tests
test:
	go test ./...

## Run tests with race detector
test-race:
	go test -race ./...

## Vet the codebase
vet:
	go vet ./...

## Run golangci-lint
lint:
	golangci-lint run ./...

## Format Go code
fmt:
	gofmt -l -w .

## Tidy modules
tidy:
	go mod tidy

## Regenerate Go protobuf bindings
proto:
	cd ../proto && buf generate
