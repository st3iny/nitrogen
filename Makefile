all: build

build:
	mkdir -p bin
	go build -ldflags="-X main.Version=$$(git rev-parse HEAD) -X main.BuildDate=$$(date -Iseconds)" -o bin/nitrogen ./cmd/nitrogen/main.go

run:
	go run ./cmd/nitrogen/main.go

test:
	go test ./internal/...

coverage:
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -func=coverage.out

.PHONY: all build run test coverage
