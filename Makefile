BINARY_NAME=hanif_skeleton

run-http:
	@go run main.go http

run-worek:
	@go run main.go worker

run-pubsub:
	@go run main.go pubsub