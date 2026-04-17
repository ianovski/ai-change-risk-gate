.PHONY: run test build fmt

run:
go run ./cmd/server

test:
go test ./...

build:
go build ./...

fmt:
gofmt -w ./cmd ./internal
