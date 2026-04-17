.PHONY: build test test-verbose test-lexer test-lexer-verbose

build:
	go build -o bin/dot ./cmd/dot/

test:
	go test ./...

test-verbose:
	go test -v ./...

test-lexer:
	go test ./internal/lexer/...

test-lexer-verbose:
	go test -v ./internal/lexer/...
