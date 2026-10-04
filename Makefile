BINARY_NAME=hanif_skeleton

run-http:
	@go run main.go http

run-worker:
	@go run main.go worker

run-pubsub:
	@go run main.go pubsub

test:
	@go test ./...

vet:
	@go vet ./...

# Project generator
.PHONY: build-cli
build-cli:
	@CGO_ENABLED=0 go build -trimpath -o bin/skeleton ./cmd/skeleton
