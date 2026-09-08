.PHONY: test lint tidy run

test:
	go test ./...

lint:
	go vet ./...

tidy:
	go mod tidy

run:
	go run ./cmd/gateway
