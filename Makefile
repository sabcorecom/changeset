MODULE := github.com/sabcorecom/changeset
BIN := bin/changeset

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)

.PHONY: build test lint vet check

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/changeset

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

check: build vet lint test
