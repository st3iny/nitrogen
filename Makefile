all: build

build:
	mkdir -p bin
	go build -o bin/nitrogen ./cmd/nitrogen/main.go

run:
	go run ./cmd/nitrogen/main.go
