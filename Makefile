export CGO_ENABLED := 0

BIN := bin/comix

.PHONY: build run

build:
	go build -o $(BIN) ./cmd

run:
	go run ./cmd
