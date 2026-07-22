all: build

setup:
	go install github.com/vladopajic/go-test-coverage/v2@latest

build:
	mkdir -p bin
	go build -ldflags="-X main.Version=$$(git rev-parse HEAD) -X main.BuildDate=$$(date -Iseconds)" -o bin/nitrogen ./cmd/nitrogen/main.go

run:
	go run ./cmd/nitrogen/main.go

test:
	go test -coverprofile=coverage.out ./internal/...

coverage: test
	go tool cover -func=coverage.out
	go-test-coverage --config=.testcoverage.yml

coverage-html: test
	go tool cover -html=coverage.out

.PHONY: all build run test coverage coverage-html
