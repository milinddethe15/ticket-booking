# Ticket Booking System Makefile

.PHONY: help build run test clean docker-build docker-run docker-down setup-db load-sample-data

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application binary"
	@echo "  run           - Run the application locally"
	@echo "  test          - Run all tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker Compose"
	@echo "  docker-down   - Stop Docker Compose services"
	@echo "  setup-db      - Setup database with migrations"
	@echo "  load-sample   - Load sample data into database"
	@echo "  fmt           - Format Go code"
	@echo "  lint          - Run linter"

# Build the application
build:
	@echo "Building application..."
	go build -o bin/ticket-booking cmd/server/main.go

# Run the application
run:
	@echo "Running application..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t ticket-booking:latest .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up --build

# Stop Docker Compose services
docker-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Setup database
setup-db:
	@echo "Setting up database..."
	docker-compose up postgres -d
	@echo "Waiting for database to be ready..."
	sleep 5
	PGPASSWORD=ticket_password psql -h localhost -U ticket_user -d ticket_booking -f migrations/001_initial_schema.up.sql

# Load sample data
load-sample:
	@echo "Loading sample data..."
	PGPASSWORD=ticket_password psql -h localhost -U ticket_user -d ticket_booking -f scripts/sample_data.sql

# Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Development setup
dev-setup: setup-db load-sample
	@echo "Development environment ready!"
	@echo "You can now run: make run"

# Production build
prod-build:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o bin/ticket-booking cmd/server/main.go

# Health check
health:
	@echo "Checking application health..."
	curl -f http://localhost:8080/health || (echo "Application is not healthy" && exit 1)

# Load test with curl
load-test:
	@echo "Running simple load test..."
	@for i in {1..50}; do \
		curl -s -X POST http://localhost:8080/api/v1/bookings \
			-H "Content-Type: application/json" \
			-d '{"user_id":1,"event_id":1,"quantity":1}' & \
	done; \
	wait; \
	echo "Load test completed"

# View logs (when running with Docker)
logs:
	docker-compose logs -f app 