.PHONY: build run test lint fmt migrate up down smoke

build:
	go build ./...

run:
	go run ./cmd/gateway

lint:
	go vet ./...

fmt:
	gofmt -l -w .

migrate:
	go run github.com/pressly/goose/v3/cmd/goose@v3.28.0 -dir ./migrations postgres "$(APP_POSTGRES_DSN)" up

up:
	docker compose -f ../deployments/docker-compose.yml up -d --build

down:
	docker compose -f ../deployments/docker-compose.yml down

smoke:
	../scripts/smoke.sh
