.PHONY: dev build test

# Install goreman: go install github.com/mattn/goreman@latest
dev:
	goreman start

build:
	go build ./cmd/her

test:
	go test ./...
