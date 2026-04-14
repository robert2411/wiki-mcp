.PHONY: build test lint run

VERSION ?= dev

build:
	go build -ldflags "-X main.version=$(VERSION)" -o wiki-mcp ./cmd/wiki-mcp

test:
	go test ./...

lint:
	golangci-lint run

run: build
	./wiki-mcp
