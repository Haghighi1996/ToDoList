.PHONY: run migrate-up migrate-down test vet fmt docker-up docker-down

run:
	go run ./cmd/api

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

docker-up:
	docker compose up -d postgres

docker-down:
	docker compose down
