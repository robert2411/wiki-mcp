.PHONY: build test lint run

VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	go build -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o wiki-mcp ./cmd/wiki-mcp

test:
	go test ./...

lint:
	golangci-lint run

run: build
	./wiki-mcp
