.PHONY: help run test build lambda tidy fmt vet docker-up docker-down docker-logs

help:
	@echo "Targets:"
	@echo "  run         - run the server locally (requires Go + local DynamoDB)"
	@echo "  test        - run unit tests"
	@echo "  build       - compile the server binary"
	@echo "  lambda      - compile the Lambda bootstrap binary (Linux/arm64) and zip"
	@echo "  tidy        - go mod tidy"
	@echo "  fmt vet     - format + static checks"
	@echo "  docker-up   - bring up full stack (DynamoDB + table init + backend)"
	@echo "  docker-down - stop the stack"
	@echo "  docker-logs - tail backend logs"

run:
	go run ./cmd/server

test:
	go test ./...

build:
	go build -o bin/server ./cmd/server

# Build for AWS Lambda (provided.al2023 runtime, arm64 = cheapest).
lambda:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap ./cmd/lambda
	zip -j flow-service.zip bootstrap
	rm bootstrap

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f backend
