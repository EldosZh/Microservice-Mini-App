.PHONY: run build test vet fmt docker-up docker-down
run:
	docker compose up --build
build:
	go build ./...
test:
	go test ./...
vet:
	go vet ./...
fmt:
	go fmt ./...
docker-up:
	docker compose up --build
docker-down:
	docker compose down
